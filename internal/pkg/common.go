package pkg

import (
	"context"
	"io"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/skyrocket-qy/erx"
	"github.com/skyrocket-qy/srand"
)

func Close(closer io.ReadCloser) {
	if err := closer.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close")
	}
}

func GetUserId(c context.Context) uint {
	return c.Value("userID").(uint)
}

func GenNumCode(n int) (string, error) {
	code := ""
	for range n {
		v, err := srand.Intn(10)
		if err != nil {
			return "", erx.W(err)
		}
		code += strconv.Itoa(v)
	}
	return code, nil
}

type EmailJob struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
	Text    string `json:"text"`
}

func Str(s string) *string {
	return &s
}
