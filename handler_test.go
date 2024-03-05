package main

import (
	"bytes"
	"io"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLambdaFunction(t *testing.T) {
	cmd := exec.Command("sam", "local", "start-api")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	time.Sleep(2 * time.Second)

	resp, err := http.Get("http://localhost:3000/user/000001")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(body), `"id": "000001"`)
}
