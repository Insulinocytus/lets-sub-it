package runner

const mockSourceVTT = `WEBVTT

00:00:00.000 --> 00:00:02.000
これは mock source 字幕の一行目です。

00:00:02.000 --> 00:00:04.000
これは mock source 字幕の二行目です。

00:00:04.000 --> 00:00:06.000
これは mock source 字幕の三行目です。
`

const mockTranslatedVTT = `WEBVTT

00:00:00.000 --> 00:00:02.000
这是 mock 翻译字幕第一行。

00:00:02.000 --> 00:00:04.000
这是 mock 翻译字幕第二行。

00:00:04.000 --> 00:00:06.000
这是 mock 翻译字幕第三行。
`

const mockBilingualVTT = `WEBVTT

00:00:00.000 --> 00:00:02.000
これは mock source 字幕の一行目です。
这是 mock 翻译字幕第一行。

00:00:02.000 --> 00:00:04.000
これは mock source 字幕の二行目です。
这是 mock 翻译字幕第二行。

00:00:04.000 --> 00:00:06.000
これは mock source 字幕の三行目です。
这是 mock 翻译字幕第三行。
`
