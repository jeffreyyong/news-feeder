package domain

type Medium string

const (
	MediumEmail   Medium = "email"
	MediumTwitter Medium = "twitter"
)

var SupportedMedium = map[Medium]bool{
	MediumTwitter: true,
}
