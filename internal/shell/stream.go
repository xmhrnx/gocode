package shell

import (
	"os"
	"os/exec"
	"time"
)

// pipeDrainGrace is how long output is still collected after the shell
// exits. A background job ("server &") keeps the pipe open indefinitely;
// the block must still finish, so its later output is dropped.
const pipeDrainGrace = 300 * time.Millisecond

// runStreaming starts cmd with stdout and stderr on one pipe and forwards
// filtered output to onOutput, on this goroutine, until the process has
// exited and the pipe is drained. It returns cmd.Wait's error.
func runStreaming(cmd *exec.Cmd, onOutput func(string)) error {
	reader, writer, err := os.Pipe()
	if err != nil {
		return err
	}
	cmd.Stdout, cmd.Stderr = writer, writer
	startErr := cmd.Start()
	// The child holds its own copy; closing ours lets EOF arrive when the
	// child and its descendants are done.
	writer.Close()
	if startErr != nil {
		reader.Close()
		return startErr
	}
	chunks, stop := readChunks(reader)
	defer close(stop)
	waited := make(chan error, 1)
	go func() { waited <- cmd.Wait() }()

	var filter outputFilter
	emit := func(text string) {
		if text != "" && onOutput != nil {
			onOutput(text)
		}
	}
	var waitErr error
	var grace <-chan time.Time
	for chunks != nil || waited != nil {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				chunks = nil
				continue
			}
			emit(filter.write(chunk))
		case waitErr = <-waited:
			waited = nil
			grace = time.After(pipeDrainGrace)
		case <-grace:
			chunks = nil
		}
	}
	emit(filter.flush())
	return waitErr
}

// readChunks reads r on its own goroutine, sending each chunk on the
// returned channel, which is closed at EOF. Closing stop makes the reader
// give up (and close r) at its next chunk instead of blocking forever.
func readChunks(r *os.File) (<-chan []byte, chan struct{}) {
	chunks := make(chan []byte)
	stop := make(chan struct{})
	go func() {
		defer r.Close()
		defer close(chunks)
		for {
			buf := make([]byte, 32<<10)
			n, err := r.Read(buf)
			if n > 0 {
				select {
				case chunks <- buf[:n]:
				case <-stop:
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()
	return chunks, stop
}
