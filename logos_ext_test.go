package logos_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.bbkane.com/gocolor"
	"go.bbkane.com/logos"
)

func TestLogger(t *testing.T) {

	goldenTest(t, goldenTestParams{
		TmpFilePrefix: "logos-test",
		FileNames:     []string{"log.jsonl", "stderr.txt", "stdout.txt"},
		GoldenDir:     filepath.Join("testdata", t.Name()),
		UpdateEnvVar:  "LOGOS_TEST_UPDATE_GOLDEN",
		WorkFunc: func(files map[string]*os.File) {
			logTmpFile := files["log.jsonl"]
			stderrTmpFile := files["stderr.txt"]
			stdoutTmpFile := files["stdout.txt"]

			zapLogger := logos.NewDeterministicZapLogger(logTmpFile)

			color, err := gocolor.Prepare(true)
			require.Nil(t, err)

			logger := logos.New(zapLogger, color, logos.WithStderr(stderrTmpFile), logos.WithStdout(stdoutTmpFile))

			logger.Debugw("debug message", "key", "value")
			logger.Infow("info message", "key", "value")
			logger.Errorw("error message", "key", "value")

			err = logger.Sync()
			require.Nil(t, err)
		},
	})
}
