package relay_test

import (
	"reflect"
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
	_, err := client.SelectResult(success, errors, false)

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
	_, err := client.SelectResult(success, errors, false)

	// This should not return an error.
	if err == nil {
		t.Errorf("Unexpected success %v\n", err)
	}
}

func TestClient_testSelectResult_noResult(t *testing.T) {
	_, client := NewDummyClient()
	empty := make([]relay.ForwardResult, 0)

	_, err := client.SelectResult(empty, empty, false)
	if err == nil {
		t.Errorf("Failed with no error %v\n", err)
	}
}

func TestClient_testSelectResult_highestChosen(t *testing.T) {
	options, client := NewDummyClient()
	var versionFieldName = "ver"
	options.VersionFieldName = &versionFieldName

	result := make([]relay.ForwardResult, 3)
	result[0].SetResultForTesting([]interface{}{
		[]uint8{'f'}, int64(3),
		[]uint8{'v', 'e', 'r'}, int64(1)})
	result[1].SetResultForTesting([]interface{}{
		[]uint8{'f'}, int64(4),
		[]uint8{'v', 'e', 'r'}, int64(3)})
	result[2].SetResultForTesting([]interface{}{
		[]uint8{'f'}, int64(4),
		[]uint8{'v', 'e', 'r'}, int64(2)})
	empty := make([]relay.ForwardResult, 0)

	res, err := client.SelectResult(result, empty, true)
	if err != nil {
		t.Errorf("No result chosen")
	}

	if !reflect.DeepEqual(res, result[1].GetResultForTesting()) {
		t.Errorf("The result is not expected: %v vs. %v\n", res,
			result[1].GetResultForTesting())
	}
}
