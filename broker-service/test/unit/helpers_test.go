package unit

// import (
// 	"broker/api"
// 	"bytes"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestReadJSON(t *testing.T) {
// 	app := &api.Config{}

// 	payload := `{"action":"auth","auth":{"email":"test@example.com","password":"password123"}}`
// 	req := httptest.NewRequest(http.MethodPost, "/handle", bytes.NewBuffer([]byte(payload)))
// 	w := httptest.NewRecorder()

// 	var requestPayload api.RequestPayload
// 	err := app.ReadJSON(w, req, &requestPayload)

// 	assert.NoError(t, err)
// 	assert.Equal(t, "auth", requestPayload.Action)
// }
