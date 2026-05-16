package core

type EventDB string

const (
	EventDBUpdated EventDB = "XKCD DB has been updated"
	EventDBDrop    EventDB = "XKCD DB has been drop"
)

type Comics struct {
	ID             int
	Title          string
	URL            string
	Description    string
	ImgDescription string `db:"img_description"`
}

type ComicsAnswer struct {
	Comics []Comics
	Total  int
}

type ComicsFWords struct {
	ID   int
	Word string
}

type ComicsSort struct {
	ID    int
	Word  string
	Score int
}
