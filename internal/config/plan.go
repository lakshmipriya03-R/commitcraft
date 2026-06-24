package config

type RewritePlan struct {
	Rewrites []RewriteItem `yaml:"rewrites"`
}

type RewriteItem struct {
	Hash        string `yaml:"hash"`
	Message     string `yaml:"message"`
	AuthorName  string `yaml:"author_name"`
	AuthorEmail string `yaml:"author_email"`
	Timestamp   string `yaml:"timestamp"`
}
