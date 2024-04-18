package frame

import (
	"testing"
)

func Test_getFrameHtml(t *testing.T) {
	type args struct {
		frameMd FrameMetaData
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "Prompt Frame",
			args: args{
				frameMd: FrameMetaData{
					Image: "imageUrl",
					Buttons: []FrameButton{
						{
							Label:  "make your own",
							Action: "link",
							Target: "luciddrops.xyz",
						},
					},
					PostURL: "post_url",
				},
			},
			want: `
			<!DOCTYPE html>
			<html>
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<meta name="description" content="luciddrops.xyz">
				<meta property="og:image" content="imageUrl">
				<meta property="fc:frame" content="vNext" />
				<meta property="fc:frame:image" content="imageUrl" />
				<meta property="fc:frame:image:aspect_ratio" content="1:1" />
				<meta property="fc:frame:button:1" content="make your own" />
				<meta property="fc:frame:button:1:action" content="link" />
				<meta property="fc:frame:button:1:target" content="luciddrops.xyz" />
			</head>
			</html>
			`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFrameHtml(tt.args.frameMd); got != tt.want {
				t.Errorf("getFrameHtml() = %v, want %v", got, tt.want)
			}
		})
	}
}
