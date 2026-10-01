package empty

import (
	"env/cli"
	"env/settings"
	"env/types"
	"env/utils"
)

type Empty struct {

	Parent   cli.Parent
	Settings *settings.Settings

	Sizes    types.Dimensions

	Label    string
	Lines    [][]string
}

func CreateEmpty(parent cli.Parent) *Empty {

	e := Empty{

		Parent:   parent,
		Settings: parent.GetSettings(),

		Sizes:    types.Dimensions{},

		Label:    "         ‹b -prsnl.spc-›b ",
		Lines:    [][]string{},
	}

	e.loadFrames()

	return &e

}

func (e *Empty) Resize(sizes types.Dimensions) {

	e.Sizes = sizes
}

func (e *Empty) Draw() *types.Queue {

	width, height := 0, 0
	for _, frame := range e.Lines {
		for _, line := range frame {
			width = max(width, utils.VisibleLength(line))
		}
		height = max(height, len(frame))
	}

	sizes := e.Sizes
	sizes.Pos[0] += max(0, (e.Sizes.Size[0]-width)/2)
	sizes.Pos[1] += max(0, (e.Sizes.Size[1]-height)/2)
	sizes.Size[0] = width
	sizes.Size[1] = height

	labelsizes := sizes
	labelsizes.Pos[0] = sizes.Pos[0]
	labelsizes.Pos[1] = sizes.Pos[1] + sizes.Size[1]
	labelsizes.Size[0] = len(e.Label)
	labelsizes.Size[1] = 1

	labelFrame := e.Settings.GenerateFrame(labelsizes, []string{e.Label}, 400, []int{0, 0, 0, 0})

	frames := []*types.Frame{}
	for _, line := range e.Lines {

		skullFrame := e.Settings.GenerateFrame(sizes, line, 120, []int{0, 0, 0, 0})

		finalFrame := e.Settings.MergeFrames(skullFrame, labelFrame)

		frames = append(frames, finalFrame)
	}

	return e.Settings.GenerateQueue(e.Sizes.Full, frames, true)
}

func (e *Empty) loadFrames() {
	e.Lines = [][]string{
		{
			"            _,,._",
			"         ,d$$$$$SIi:.",
			"       ,$$$SSSS$$SSIi:.",
			"      j$$$$SSSS$$$SIIi:.",
			"     .S$$$$SS$$$$$SSIi:.",
			"     j?ᵒ`‾`?S$SI7ᵒ\"ᵃ?IL:.",
			"     ?:     $$S?     `?i'",
			"     iL    j$?$k.     I7",
			"     $$$b%d$'  `$b,_.dS:",
			"     ?SSIiS?    S$?I?$Si",
			"      ‾`?IS$L_,d$SIi:`ᵃ'",
			"         ?$$$SS$SIi'",
			"         j:?i:i?.:·:",
			"           \"` `^",
		},
		{
			"            _.,,._",
			"       _,oS$$$$$SSIi:,_",
			"     ,d$$$$$$$S$$$SSISi:.",
			"    ,$$$$$SSSS$$$$SIISSIi:.",
			"   j$$$$SS$$$$$$SIISS$$SIi:",
			"  .\"I7‾`ᵒᵃ$$iI$SiIS$$$$SIi:",
			"  `.jI    `?L:iIIS$$$$SII:.",
			"  ,ᵃ$$.    j$b:iIS$$S?iIi:.",
			"  ? i$L,_.d$d$:ijSSSIiSi:.",
			"  i.j$S?S$$$$S%u,ᵒ?iISi::.",
			" ,d$$SIi?ᵃ?$S?ᵃ`   ‾`‾.:'",
			" i$$SI$$k.         ..'",
			" :?::.i?·^",
			" ",
		},
		{
			"            _.,,,._",
			"       .,uS$$$$$$$SSi:,_",
			"     ,d$$S$$$$S$$$SS$SIi:.",
			"    j$S$$S7IS$$S$$$$S$SIi::",
			"   d$$$7iIS$$$$$$SS$$SIIi:.",
			"  $$$7:iIS$S$$$SIS$$$$SIi::",
			"  j ?k:iISIS$S&IS$$$SSIIi::",
			".d? j$b:iIj$$S7IIS$$S7IIi::.",
			"`$$u$$7.:j$$SiiIS$$$SIIIi::.",
			" biS$S$p,._?%u,`?S$SSIIi::.",
			"d$$$Si?ᵃ?$S?ᵃ\"`^\"4ISIii::.",
			"ᵃ?$$k.           .?7I:.",
			"^:?i:\"`          .·'‾",
			" ",
		},
		{
			"             _.,,,._",
			"        .,oS$$$$$$$SSi:,_",
			"      ,d$$S$$$$S$$$$SSSIi:.",
			"     j$S$$$$$S$$S$$$$$$Sli::",
			"    j$$$$$$SS$$$$$$$S$$SSII:.",
			"   j$7iIS$$$$S$$$$$$IS$$SIi::",
			"  d$7:jS$SIiIS$$$$$SS$$$Si::.",
			"  ?7.j$$Siid$$$$$$SS$$$S7i::.",
			"  j:.?$liid$$$$SIS$$$SSSIi::.",
			"  $k,_`$IS$$$$SIIS$$$SSIii:.",
			"  ?'  ‾`ᵒᵃ$S?ᵃ::::iISSii::.",
			"   k.·:'   i7   ··::::::··",
			"   ``      \"'`",
			" ",
		},
		{
			"            _.,,._",
			"         ,d$$$$$SSi:,",
			"     a,d$S$$$$$$$$SSIi:.",
			"     dS$$$$$S$$S$$$$$SIk.",
			"    :S$$$$SS$$$$$$$$$$SIi",
			"    IS$$$$$S$$$$$$$$$SIi:",
			"    SS$$$$$7S$$$$$S$$SLi·",
			"    ?$$$$$7jSS$$$$S$SIi:",
			"    ji$$S&j$Si?S$$SSIi:·",
			"    ?$IuiI$$$Ski?S7Ii:·'",
			"     ?\\Sbp.`ᵒᵃ?S$7':i:·",
			"     ? \"‾ ·:::. .::..",
			" ",
			" ",
		},
		{
			"            _.,,._",
			"      ,  uS$$$$$SSIi:,",
			"   ,  d$$S$$$$$$$$$$SIi:.",
			"    dIS$$$$$$$$S$$$$$SSSik",
			"   jIS$$$$$$$$$$$$$$$$SIiiL",
			"  ·IIS$$$$$S$$$$$$$$$$SI:?$",
			"  :iS$$$$$$7S$$$$SS$$SIii:?k",
			"  :iS$$$$$7jIS$$$$SS$SIS::·?",
			"  ·iIS$$S7j$SI?S$$$SSii7 · L",
			"   :iS$SSi$$$SL`?S$SIi?_·o$$",
			"    ?ISi7 `ᵒᵃ?Sb,`ᵃᵒ'‾`  _`\"",
			"     ᵃ?ᵃ'··:::·`?S$i'  · :",
			"             ··  `?'",
			" ",
		},
		{
			"            _.,,._",
			"       _,dS$$$$$SSIi:,",
			"     ,dS$S$$$$$$$$SSSIi:.",
			"    dIS$$$$$$$$S$$$$SSSSIk",
			"   dIS$$$$$$$$$$$$$$$SSSiiL",
			"  iLSS$$$$SS$$$$$$$$$SSIi?$k",
			"  SiSS$$S$SSS$$$$S$$SSIIi:S?",
			"  iiS$$$$ISSIS$$$$SSSIIi?·j.",
			"  :iIS$SIIS$SI?S$$SIIi i7'jI$:",
			"   :iISSiiiS$SLi?SI ?ᵃ',od$S$",
			"    ·:iSi:?S$SI?'ᵃᵒ'‾`^ᵒᵃ?Sk",
			"      `ᵒᵃ.:?S$Si      ·::iI$$",
			"            `ᵒᵃ'        \".ᵃ:'",
			" ",
		},
		{
			"            _,,,._",
			"        _,d$$$$$$S$Sik,",
			"      ,i$$S$$$$$$$$SSSIk:",
			"     dISS$$$$$$$SS$$$$SSIk",
			"    jIS$$$$$$$$$SIIS$$$SSIk",
			"    SSS$$$$SS$$SSIIi$$S7ᵃ??k",
			"    ?SS$$$$$$SSSIi:d$'   :?",
			"    :SS$$$$$SSIi::j$S    j$7",
			"     ?IS$$SSIi::.,?$$$up%?$'",
			"      ?ISIb,‾`ᵒ\"?S$$ᵃᵒ‾,d$'",
			"       `ᵒ?$?'    `?ᵃk.:iIS$$k.",
			"                    `?.:iI$$i",
			"                     \".\"ᵃ ^ '",
			" ",
		},
		{
			"            _.,,._",
			"        _,d$$$$$$$$b.",
			"      ,d$S$$$$S$$$$SSb.",
			"     dSS$SIS$$$$$$$$$Sib",
			"    jIS$$SII$$$$$$$$$S$Sk",
			"    SI$$IIiid$$S?iI$$$S7ᵃk.",
			"   :iS$IIi:j7'     ?$S?   i",
			"   .iSSI:::$$      j$$L   ?",
			"    :?i:.,d$$k,_.,d$$7‾?p,$",
			"     iI:`ᵃ?$$$SS?I?$$7  $$?",
			"      ?:.  ?S$$SII$$L.,J$\"'",
			"       `.   ‾  :IS$S$$$Sk",
			"               .:i?i:?:.?",
			"                    \"` `",
		},
	}
}
