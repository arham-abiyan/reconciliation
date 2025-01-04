package pkg

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ValidateDates(startDate, endDate string) error {
	if startDate == "" || endDate == "" {
		return fmt.Errorf("both start_date and end_date are required")
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return fmt.Errorf("invalid start_date format: use YYYY-MM-DD")
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Errorf("invalid end_date format: use YYYY-MM-DD")
	}

	if end.Before(start) {
		return fmt.Errorf("end_date cannot be before start_date")
	}

	return nil
}

func ValidateFile(fileHeader *multipart.FileHeader) error {
	if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".csv") {
		return fmt.Errorf("invalid file type")
	}

	return nil
}

func SaveFile(fileHeader *multipart.FileHeader, uploadsDir, prefix string) (string, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Create unique filename with timestamp
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s_%s", prefix, timestamp, fileHeader.Filename)
	dst, err := os.Create(filepath.Join(uploadsDir, filename))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", uploadsDir, filename), nil
}
