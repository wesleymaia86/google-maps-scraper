package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type geocodeResponse struct {
	Name    string  `json:"name"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Display string  `json:"display"`
}

type latlngFeatureCollection struct {
	Features []latlngFeature `json:"features"`
}

type latlngFeature struct {
	Geometry struct {
		Coordinates []float64 `json:"coordinates"`
	} `json:"geometry"`
	Properties struct {
		Name    string `json:"name"`
		State   string `json:"state"`
		Country string `json:"country"`
		Type    string `json:"type"`
	} `json:"properties"`
}

func (s *Server) apiGeocode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		renderJSON(w, http.StatusMethodNotAllowed, apiError{
			Code:    http.StatusMethodNotAllowed,
			Message: "Método não permitido",
		})

		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		renderJSON(w, http.StatusUnprocessableEntity, apiError{
			Code:    http.StatusUnprocessableEntity,
			Message: "Informe a cidade ou bairro",
		})

		return
	}

	apiKey := os.Getenv("LATLNG_API_KEY")
	if apiKey == "" {
		renderJSON(w, http.StatusServiceUnavailable, apiError{
			Code:    http.StatusServiceUnavailable,
			Message: "Geocodificação não configurada (LATLNG_API_KEY)",
		})

		return
	}

	// Bias Brazil when the query has no country hint.
	if !strings.Contains(strings.ToLower(q), "brasil") && !strings.Contains(strings.ToLower(q), "brazil") {
		q = q + ", Brasil"
	}

	endpoint := "https://api.latlng.work/api?" + url.Values{
		"q":     {q},
		"limit": {"5"},
	}.Encode()

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, apiError{
			Code:    http.StatusInternalServerError,
			Message: "Falha ao montar a busca de localização",
		})

		return
	}

	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		renderJSON(w, http.StatusBadGateway, apiError{
			Code:    http.StatusBadGateway,
			Message: "Falha ao consultar o serviço de geocodificação",
		})

		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		renderJSON(w, http.StatusBadGateway, apiError{
			Code:    http.StatusBadGateway,
			Message: "Resposta inválida do serviço de geocodificação",
		})

		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		renderJSON(w, http.StatusBadGateway, apiError{
			Code:    http.StatusBadGateway,
			Message: "Chave da API de geocodificação inválida",
		})

		return
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		renderJSON(w, http.StatusTooManyRequests, apiError{
			Code:    http.StatusTooManyRequests,
			Message: "Limite de geocodificação atingido. Tente novamente mais tarde",
		})

		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		renderJSON(w, http.StatusBadGateway, apiError{
			Code:    http.StatusBadGateway,
			Message: fmt.Sprintf("Geocodificação retornou status %d", resp.StatusCode),
		})

		return
	}

	var collection latlngFeatureCollection
	if err := json.Unmarshal(body, &collection); err != nil {
		renderJSON(w, http.StatusBadGateway, apiError{
			Code:    http.StatusBadGateway,
			Message: "Não foi possível interpretar a resposta de geocodificação",
		})

		return
	}

	if len(collection.Features) == 0 {
		renderJSON(w, http.StatusNotFound, apiError{
			Code:    http.StatusNotFound,
			Message: "Nenhuma localização encontrada para essa busca",
		})

		return
	}

	feature := collection.Features[0]
	if len(feature.Geometry.Coordinates) < 2 {
		renderJSON(w, http.StatusBadGateway, apiError{
			Code:    http.StatusBadGateway,
			Message: "Coordenadas ausentes na resposta",
		})

		return
	}

	// GeoJSON order is [longitude, latitude].
	lon := feature.Geometry.Coordinates[0]
	lat := feature.Geometry.Coordinates[1]

	parts := []string{feature.Properties.Name}
	if feature.Properties.State != "" {
		parts = append(parts, feature.Properties.State)
	}
	if feature.Properties.Country != "" {
		parts = append(parts, feature.Properties.Country)
	}

	name := feature.Properties.Name
	if name == "" {
		name = q
	}

	renderJSON(w, http.StatusOK, geocodeResponse{
		Name:    name,
		Lat:     lat,
		Lon:     lon,
		Display: strings.Join(parts, ", "),
	})
}
