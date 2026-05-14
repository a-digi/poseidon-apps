package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// setupLogging routes stdlib log output to both stdout (so longrunning.Registry
// captures it) and ${dataDir}/coco-gg.log (so the user can tail -f even when
// the host UI doesn't expose stdout). The returned io.Closer is the file
// handle; the caller defer-closes it.
//
// O_TRUNC: each start is a fresh log window. The port file uses the same
// convention; avoids unbounded growth in long sessions.
func setupLogging(dataDir string) (io.Closer, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(dataDir, "coco-gg.log"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}
	log.SetOutput(io.MultiWriter(os.Stdout, f))
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetPrefix("[coco-gg] ")
	return f, nil
}
