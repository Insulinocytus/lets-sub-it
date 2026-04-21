package pipeline

type Segment struct {
	Start          string `json:"start"`
	End            string `json:"end"`
	SourceText     string `json:"sourceText"`
	TranslatedText string `json:"translatedText"`
}
