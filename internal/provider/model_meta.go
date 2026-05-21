package provider

// ModelMeta 是模型元信息。
type ModelMeta struct {
	MaxContextTokens int
	MaxOutputTokens  int
	SupportsStream   bool
}
