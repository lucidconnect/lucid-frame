package frame

import (
	"fmt"
	"strings"
)

type FrameMetaData struct {
	Image   string
	Buttons []FrameButton
	PostURL string
}

type FrameButton struct {
	Label  string
	Action string
	Target string
}

func getFrameHtml(frameMd FrameMetaData) string {
	tags := getBaseMetaTags()

	// compose button tags
	buttonTags := []string{
		fmt.Sprintf(`<meta property="og:image" content="%v">`, frameMd.Image),
		fmt.Sprintf(`<meta property="fc:frame:image" content="%v" />`, frameMd.Image),
	}
	for i, v := range frameMd.Buttons {
		buttonIdx := i + 1
		buttonTags = append(buttonTags,
			fmt.Sprintf(`<meta property="fc:frame:button:%v" content="%v" />`, buttonIdx, v.Label),
			fmt.Sprintf(`<meta property="fc:frame:button:%v:action" content="%v" />`, buttonIdx, v.Action),
		)

		if v.Target != "" {
			buttonTags = append(buttonTags, fmt.Sprintf(`<meta property="fc:frame:button:%v:target" content="%v" />`, buttonIdx, v.Target))
		}
	}

	if frameMd.PostURL != "" {
		buttonTags = append(buttonTags, fmt.Sprintf(`<meta property="fc:frame:post_url" content="%v" />`, frameMd.PostURL))
	}

	tags = append(tags, buttonTags...)
	tags = append(tags, `</head>`)
	tags = append(tags, `</html>`)

	return strings.Join(tags, "")
}

// func returnMintFrame() {

// }

func getBaseMetaTags() []string {
	tags := []string{
		`<!DOCTYPE html>`,
		`<html>`,
		`<head>`,
		`<meta charset="UTF-8">`,
		`<meta name="viewport" content="width=device-width, initial-scale=1.0">`,
		`<meta name="description" content="luciddrops.xyz">`,
		`<meta property="fc:frame:image:aspect_ratio" content="1:1" />`,
		`<meta property="fc:frame" content="vNext" />`,
	}

	return tags
}
