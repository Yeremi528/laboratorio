package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Yeremi528/laboratorio/foundation/mask"
)

// Respond converts the input data to JSON and sends it to the client.
func Respond(ctx context.Context, w http.ResponseWriter, data any, statusCode int) error {
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	SetStatusCode(ctx, statusCode)

	maskedResponse, err := mask.StructToByte(data)
	if err != nil {
		return nil
	}

	SetResponse(ctx, string(maskedResponse))

	return nil
}
