package hello

import "github.com/emersonjoe/trilha"

// GET /api/hello
func GET(c *trilha.Ctx) error {
	return c.JSON(200, map[string]string{"hello": "wavehub-voice", "request_id": c.RequestID()})
}
