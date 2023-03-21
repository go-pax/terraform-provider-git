package git

type File struct {
	Contents string `json:"contents"`
	FilePath string `json:"filepath"`
}
