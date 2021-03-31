package install

import (
	"io"
	"os"
	"os/exec"
	"sync"
)

// Commander is a wrapper of exec.Command
type Commander struct {
	Env []string
}

func (c Commander) execCommand(name string, args ...string) (err error) {
	command := exec.Command(name, args...)
	if len(c.Env) > 0 {
		command.Env = append(command.Env, c.Env...)
	}

	//var stdout []byte
	//var errStdout error
	stdoutIn, _ := command.StdoutPipe()
	stderrIn, _ := command.StderrPipe()
	err = command.Start()
	if err != nil {
		return err
	}

	// cmd.Wait() should be called only after we finish reading
	// from stdoutIn and stderrIn.
	// wg ensures that we finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_, _ = c.copyAndCapture(os.Stdout, stdoutIn)
		wg.Done()
	}()

	_, _ = c.copyAndCapture(os.Stderr, stderrIn)

	wg.Wait()

	err = command.Wait()
	return
}

func (c Commander) copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}
