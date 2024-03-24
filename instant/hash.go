package instant

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// HashType is an answer Type
const HashType Type = "hash"

// Hash is an instant answer
type Hash struct {
	Answer
}

// HashAlgo is a hashing algorithm
type HashAlgo string

const (
	// MD5 is the md5 algorithm
	MD5 HashAlgo = "MD5"
	// SHA1 is the sha1 algorithm
	SHA1 HashAlgo = "SHA1"
	// SHA224 is the sha224 algorithm
	SHA224 HashAlgo = "SHA224"
	// SHA256 is the sha256 algorithm
	SHA256 HashAlgo = "SHA256"
	// SHA512 is the sha512 algorithm
	SHA512 HashAlgo = "SHA512"
)

// HashResponse is the response to the instant answer
type HashResponse struct {
	Original string
	HashAlgo
	Solution string
}

func (h *Hash) setQuery(r *http.Request, qv string) Answerer {
	h.Answer.setQuery(r, qv)
	return h
}

func (h *Hash) setUserAgent(r *http.Request) Answerer {
	return h
}

func (h *Hash) setLanguage(lang language.Tag) Answerer {
	h.language = lang
	return h
}

func (h *Hash) setType() Answerer {
	h.Type = HashType
	return h
}

func (h *Hash) setRegex() Answerer {
	triggers := []string{
		"md5", "sha", "sha1", "sha224", "sha256", "sha512",
	}

	t := strings.Join(triggers, "|")
	h.regex = append(h.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))

	return h
}

func (h *Hash) solve(r *http.Request) Answerer {
	h.remainder = strings.TrimPrefix(h.remainder, "hash ")
	h.remainder = strings.TrimPrefix(h.remainder, "of ")
	h.remainder = strings.TrimPrefix(h.remainder, `"`)
	h.remainder = strings.TrimSuffix(h.remainder, `"`)

	d := []byte(strings.TrimSpace(h.remainder))

	sol := HashResponse{
		Original: h.remainder,
	}

	switch h.triggerWord {
	case "md5":
		sol.HashAlgo = MD5
		sol.Solution = fmt.Sprintf("%x", md5.Sum(d))
	case "sha", "sha1":
		sol.HashAlgo = SHA1
		sol.Solution = fmt.Sprintf("%x", sha1.Sum(d))
	case "sha224":
		sol.HashAlgo = SHA224
		sol.Solution = fmt.Sprintf("%x", sha256.Sum224(d))
	case "sha256":
		sol.HashAlgo = SHA256
		sol.Solution = fmt.Sprintf("%x", sha256.Sum256(d))
	case "sha512":
		sol.HashAlgo = SHA512
		sol.Solution = fmt.Sprintf("%x", sha512.Sum512(d))
	default:
		h.Err = fmt.Errorf("unknown hashing algorigthm %v", h.triggerWord)
		return h
	}

	h.Solution = sol

	return h
}

func (h *Hash) tests() []test {
	tests := []test{
		{
			query: "md5 this",
			expected: []Data{
				{
					Type:      HashType,
					Triggered: true,
					Solution: HashResponse{
						Original: "this",
						HashAlgo: MD5,
						Solution: "9e925e9341b490bfd3b4c4ca3b0c1ef2",
					},
				},
			},
		},
		{
			query: `sha hash of "this entire string"`,
			expected: []Data{
				{
					Type:      HashType,
					Triggered: true,
					Solution: HashResponse{
						Original: "this entire string",
						HashAlgo: SHA1,
						Solution: "dd5c370a950f4dbb48a6212b0bde03eb3a021897",
					},
				},
			},
		},
		{
			query: `sha1 "this entire string"`,
			expected: []Data{
				{
					Type:      HashType,
					Triggered: true,
					Solution: HashResponse{
						Original: "this entire string",
						HashAlgo: SHA1,
						Solution: "dd5c370a950f4dbb48a6212b0bde03eb3a021897",
					},
				},
			},
		},
		{
			query: `sha224 hash of "this entire string"`,
			expected: []Data{
				{
					Type:      HashType,
					Triggered: true,
					Solution: HashResponse{
						Original: "this entire string",
						HashAlgo: SHA224,
						Solution: "f9cbc8589549f186e44921d765a93719f380097e0af88070bf6607a9",
					},
				},
			},
		},
		{
			query: `sha256 hash of "this entire string"`,
			expected: []Data{
				{
					Type:      HashType,
					Triggered: true,
					Solution: HashResponse{
						Original: "this entire string",
						HashAlgo: SHA256,
						Solution: "820b0b1b609e77038b1f37e623e7f05cce9f7727fd1f557607e9badd431d208f",
					},
				},
			},
		},
		{
			query: `sha512 of another string`,
			expected: []Data{
				{
					Type:      HashType,
					Triggered: true,
					Solution: HashResponse{
						Original: "another string",
						HashAlgo: SHA512,
						Solution: "410f7993f53b148c5b439c8e48fd5083860d648a00ff7579b0046257822c35658591bddc662ea8bda650cd729f1f3f876038240fa0422a811cc00eeff170e500",
					},
				},
			},
		},
	}

	return tests
}
