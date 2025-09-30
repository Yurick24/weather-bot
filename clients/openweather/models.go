package openweather

type CoordinatesResponse struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

type Coordinates struct {
	Lat float64
	Lon float64
}

type WeatherResponse struct {
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
		GrndLevel int     `json:"grnd_level"`
	} `json:"main"`
	Visibility int `json:"visibility"`

	Wind struct {
		Speed float64 `json:"speed"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`

	Rain struct {
		OneH float64 `json:"1h"`
	} `json:"rain"`

	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Name string `json:"name"`
}

type Weather struct {
	NameCity      string
	Temp          float64
	FeelsLike     float64
	Description   string
	Precipitation float64
	WindSpeed     float64
	WindGust      float64
	GrndLevel     int
	Humidity      int
	Visibility    int
	Clouds        int
}
