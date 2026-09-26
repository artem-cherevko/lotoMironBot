package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"log"
	"lotoMironBot/internal/services"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/go-telegram/bot"
	"github.com/redis/go-redis/v9"
)

const (
	ocrURL   = "https://paddleocr.aistudio-app.com/api/v2/ocr/jobs"
	ocrModel = "PP-OCRv6"
)

type Worker struct {
	b     *bot.Bot
	r     *redis.Client
	lotoS *services.LotoService
	token string
}

func NewWorker(b *bot.Bot, r *redis.Client, token string, lotoS *services.LotoService) *Worker {
	return &Worker{
		b:     b,
		r:     r,
		token: token,
		lotoS: lotoS,
	}
}

var (
	ocrSemaphore = make(chan struct{}, 1)
	numberRegex  = regexp.MustCompile(`\d+`)
)

func (w *Worker) Process(ctx context.Context, fileID string) ([]int32, error) {
	// Telegram GetFile
	file, err := w.b.GetFile(ctx, &bot.GetFileParams{
		FileID: fileID,
	})
	if err != nil {
		return nil, err
	}

	// Download
	link := w.b.FileDownloadLink(file)

	path := fmt.Sprintf("./images/%s.jpeg", fileID)
	croppedPath := fmt.Sprintf("./images/%s_crop.png", fileID)

	defer os.Remove(path)
	defer os.Remove(croppedPath)

	if err := os.MkdirAll("./images", 0755); err != nil {
		return nil, err
	}

	cmd := exec.Command(
		"curl",
		"-L",
		"-o",
		path,
		link,
	)

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// Crop
	img, err := imaging.Open(path)
	if err != nil {
		return nil, err
	}

	cropped := imaging.Crop(
		img,
		image.Rect(
			450,
			200,
			1250,
			550,
		),
	)

	if err := imaging.Save(cropped, croppedPath); err != nil {
		return nil, err
	}

	// OCR
	numbers, err := w.OCR(ctx, croppedPath)
	if err != nil {
		return nil, err
	}

	return numbers, nil
}

func (w *Worker) OCR(ctx context.Context, path string) ([]int32, error) {
	// Только один запрос одновременно
	ocrSemaphore <- struct{}{}
	defer func() {
		<-ocrSemaphore
	}()

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Multipart body
	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("model", ocrModel); err != nil {
		return nil, err
	}

	if err := writer.WriteField(
		"optionalPayload",
		`{
			"useDocOrientationClassify": false,
			"useDocUnwarping": false,
			"useTextlineOrientation": false
		}`,
	); err != nil {
		return nil, err
	}

	part, err := writer.CreateFormFile(
		"file",
		filepath.Base(path),
	)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}

	contentType := writer.FormDataContentType()

	if err := writer.Close(); err != nil {
		return nil, err
	}

	// Submit job
	var upload struct {
		Data struct {
			JobID string `json:"jobId"`
		} `json:"data"`
	}

	for attempt := 1; attempt <= 10; attempt++ {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			ocrURL,
			bytes.NewReader(body.Bytes()),
		)
		if err != nil {
			return nil, err
		}

		req.Header.Set(
			"Authorization",
			"bearer "+w.token,
		)

		req.Header.Set(
			"Content-Type",
			contentType,
		)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if readErr != nil {
			return nil, readErr
		}

		if resp.StatusCode == http.StatusOK {
			if err := json.Unmarshal(data, &upload); err != nil {
				return nil, err
			}

			break
		}

		// PaddleOCR queue full
		if bytes.Contains(data, []byte(`"code":10010`)) {
			log.Printf(
				"[OCR] queue full, retry %d/10",
				attempt,
			)

			if attempt == 10 {
				return nil, fmt.Errorf(
					"OCR queue full after %d attempts",
					attempt,
				)
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()

			case <-time.After(time.Duration(attempt*5) * time.Second):
			}

			continue
		}

		return nil, fmt.Errorf(
			"OCR upload: status=%d body=%s",
			resp.StatusCode,
			data,
		)
	}

	if upload.Data.JobID == "" {
		return nil, errors.New("OCR returned empty jobId")
	}

	// Polling
	var result struct {
		Data struct {
			State    string `json:"state"`
			ErrorMsg string `json:"errorMsg"`

			ResultURL struct {
				JSONURL string `json:"jsonUrl"`
			} `json:"resultUrl"`
		} `json:"data"`
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case <-time.After(2 * time.Second):
		}

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			ocrURL+"/"+upload.Data.JobID,
			nil,
		)
		if err != nil {
			return nil, err
		}

		req.Header.Set(
			"Authorization",
			"bearer "+w.token,
		)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if readErr != nil {
			return nil, readErr
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf(
				"OCR status: %d body=%s",
				resp.StatusCode,
				data,
			)
		}

		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}

		switch result.Data.State {
		case "pending":
			log.Printf(
				"[OCR] job=%s pending",
				upload.Data.JobID,
			)

		case "running":
			log.Printf(
				"[OCR] job=%s running",
				upload.Data.JobID,
			)

		case "failed":
			return nil, fmt.Errorf(
				"OCR failed: %s",
				result.Data.ErrorMsg,
			)

		case "done":
			goto DONE

		default:
			log.Printf(
				"[OCR] job=%s unknown state=%s",
				upload.Data.JobID,
				result.Data.State,
			)
		}
	}

DONE:

	if result.Data.ResultURL.JSONURL == "" {
		return nil, errors.New("OCR returned empty result URL")
	}

	// Получаем JSONL
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		result.Data.ResultURL.JSONURL,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf(
			"OCR result: status=%d body=%s",
			resp.StatusCode,
			data,
		)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Парсим числа
	var numbers []int32

	for _, line := range bytes.Split(data, []byte("\n")) {
		line = bytes.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		var r struct {
			Result struct {
				OCRResults []struct {
					PrunedResult struct {
						RecTexts []string `json:"rec_texts"`
					} `json:"prunedResult"`
				} `json:"ocrResults"`
			} `json:"result"`
		}

		if err := json.Unmarshal(line, &r); err != nil {
			return nil, err
		}

		for _, page := range r.Result.OCRResults {
			for _, text := range page.PrunedResult.RecTexts {

				matches := numberRegex.FindAllString(text, -1)

				for _, match := range matches {
					n, err := strconv.Atoi(match)
					if err != nil {
						continue
					}

					numbers = append(numbers, int32(n))
				}
			}
		}
	}

	return numbers, nil
}

func (w *Worker) Run(ctx context.Context) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		keys, err := w.r.Keys(ctx, "ticket:*").Result()
		if err != nil {
			log.Printf("[worker] redis keys: %v", err)

			select {
			case <-ctx.Done():
				return

			case <-time.After(time.Second):
			}

			continue
		}

		for _, key := range keys {
			result, err := w.r.BRPop(
				ctx,
				time.Second,
				key,
			).Result()

			if err != nil {
				continue
			}

			fileID := result[1]

			numbers, err := w.Process(ctx, fileID)
			if err != nil {
				log.Printf(
					"[worker] collection=%s file=%s error=%v",
					key,
					fileID,
					err,
				)

				continue
			}

			log.Printf(
				"[worker] collection=%s file=%s numbers=%v",
				key,
				fileID,
				numbers,
			)

			w.lotoS.AddTicket(ctx, fileID, strings.Split(key, ":")[1], numbers)
		}
	}
}
