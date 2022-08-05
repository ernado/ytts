package main

import (
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ernado/ytts"
)

func main() {
	var (
		s   ytts.Options
		out string
	)
	flag.StringVar(&s.Text, "text", "Привет, мир!", "text to synthesize")
	flag.StringVar(&s.Language, "lang", "ru-RU", "language")
	flag.StringVar(&s.Voice, "voice", "omazh", "voice")
	flag.StringVar(&s.Emotion, "emotion", "neutral", "emotion")
	flag.Float64Var(&s.Speed, "speed", -1, "speed")
	flag.StringVar(&out, "out", "", "output file (generate if not set)")
	flag.Parse()

	if out == "" {
		h := md5.New()
		_, _ = io.WriteString(h, s.Text)
		out = fmt.Sprintf("%s-%s-%x.ogg", s.Voice, s.Emotion, h.Sum(nil)[:4])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	body, err := ytts.New(os.Getenv("YANDEX_TOKEN"),
		ytts.WithFolderID(os.Getenv("YANDEX_FOLDER_ID")),
	).Synthesize(ctx, s)
	if err != nil {
		panic(err)
	}
	defer func() { _ = body.Close() }()
	f, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	if _, err := io.Copy(f, body); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	fmt.Println("Saved to", out)
}
