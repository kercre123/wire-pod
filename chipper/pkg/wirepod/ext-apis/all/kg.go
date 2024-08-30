package extapis

type KGAPI struct {
	Name         string
	Link         string
	NeedsPayment bool
	APIAddr      string
	// streaming capable, as well
	IGCapable bool
}

type WeatherAPI struct {
	Name          string
	Link          string
	NeedsPayment  bool
	APIAddr       string
	GeoAddr       string
	NeedsGeo      bool
	Structure     interface{}
	GeoStructure  interface{}
	Meteorologist WeatherAPIer
}

type WeatherAPIer interface {
	/*
		functions you need to define
	*/
	GetWeatherWithLocation(string) WeatherConditions
	GetWeatherWithCoordinates(Coordinates) WeatherConditions
	GetCoordinates() Coordinates
	// test the API, see if API key is correct
	Test() (bool, error)

	/*
		functions you DON'T need to define
	*/
	// actual entry-point function. this is already defined
	// if NeedsGeo is false, it won't use GetWeatherWithCoordinates, vice versa
	GetWeather(string) WeatherConditions
}
