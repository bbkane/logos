package logos_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.bbkane.com/gocolor"
	"go.bbkane.com/logos"
)

// goldenTest provides files to work(), then compares those files to previously
// saved ones and provides vimdiff commands to inspect any diffences found.
// In pseudocode for a one-file version of this function:
//
//	tmpFile := NewTmpFile()
//	Write(tmpFile)  // work
//	Close(tmpFile)
//	actualBytes := Read(tmpFile.Name())
//	var goldenFilePath
//	if update {
//	    Write(actualBytes, goldenFilePath)
//	}
//	expectedBytes := Read(goldenFilePath)
//	Compare(expectedBytes, actualBytes)
func goldenTest(
	t *testing.T,
	tmpFilePrefix string,
	names []string,
	goldenDir string,
	update bool,
	work func(map[string]*os.File),
) {
	var tmpFiles = make(map[string]*os.File, len(names))
	for _, name := range names {
		file, err := os.CreateTemp(os.TempDir(), tmpFilePrefix+"-"+name)
		require.Nil(t, err)
		t.Logf("wrote tmpfile: %#v", file.Name())
		tmpFiles[name] = file
	}

	work(tmpFiles)

	for _, name := range names {
		tmpFile := tmpFiles[name]
		err := tmpFile.Close()
		require.Nil(t, err)

		actualBytes, err := os.ReadFile(tmpFile.Name())
		require.Nil(t, err)

		goldenFilePath := filepath.Join(goldenDir, name+".golden.txt")
		goldenFilePath, err = filepath.Abs(goldenFilePath)
		require.Nil(t, err)

		if update {
			err = os.MkdirAll(goldenDir, 0700)
			require.Nil(t, err)

			err = os.WriteFile(goldenFilePath, actualBytes, 0600)
			require.Nil(t, err)
			t.Logf("wrote golden file: %#v\n", goldenFilePath)
		}

		expectedBytes, err := os.ReadFile(goldenFilePath)
		require.Nil(t, err)

		if !bytes.Equal(expectedBytes, actualBytes) {
			t.Logf(
				"%s: expected != actual. See diff:\n  vimdiff %s %s\n\n",
				name,
				goldenFilePath,
				tmpFile.Name(),
			)
			t.Fail()
		}
	}
}

func TestLogger(t *testing.T) {

	update := os.Getenv("LOGOS_TEST_UPDATE_GOLDEN") != ""

	goldenTest(
		t,
		"logos-test",
		[]string{"log", "stderr", "stdout"},
		filepath.Join("testdata", t.Name()),
		update,
		func(files map[string]*os.File) {
			logTmpFile := files["log"]
			stderrTmpFile := files["stderr"]
			stdoutTmpFile := files["stdout"]

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
	)
}
