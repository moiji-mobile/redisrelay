package relay_test

import (
	"testing"

	"github.com/moiji-mobile/redisrelay/relay"
	"go.uber.org/zap"
)

func NewDummyClient() (*relay.ServerOptions, *relay.Client) {
	// Create an option and bind to a random port
	options := relay.DefaultOptions()
	logger, _ := zap.NewDevelopment()

	// The final client
	client := relay.NewClient(&options, logger)

	// And its returned
	return &options, client
}

func TestClient_testSelectResult_success(t *testing.T) {
	options, client := NewDummyClient()
	var minSuccess = uint32(2)
	options.MinSuccess = &minSuccess

	errors := make([]relay.ForwardResult, 10)
	success := make([]relay.ForwardResult, 2)
	_, err := client.SelectResult(success, errors)

	// This should not return an error.
	if err != nil {
		t.Errorf("Unexpected error %v\n", err)
	}
}

func TestClient_testSelectResult_notEnoughSuccess(t *testing.T) {
	options, client := NewDummyClient()
	var minSuccess = uint32(2)
	options.MinSuccess = &minSuccess

	errors := make([]relay.ForwardResult, 0)
	success := make([]relay.ForwardResult, 1)
	_, err := client.SelectResult(success, errors)

	// This should not return an error.
	if err == nil {
		t.Errorf("Unexpected success %v\n", err)
	}
}

func TestClient_testSelectResult_noResult(t *testing.T) {
	_, client := NewDummyClient()
	empty := make([]relay.ForwardResult, 0)

	_, err := client.SelectResult(empty, empty)
	if err == nil {
		t.Errorf("Failed with no error %v\n", err)
	}
}
