package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type gui struct {
	a fyne.App
	w fyne.Window
}

type channel struct {
	Name string
	// Pic, Info...
	Header   string
	Messages []message
}

type account struct {
	Name string
	Desc string
}

type message struct {
	Date    time.Time // when it was posted
	Poster  account
	Message string
	Image   fyne.Resource
}

func main() {
	g := gui{}

	g.a = app.New()
	g.w = g.a.NewWindow("Billions of Messages")
	g.w.Resize(fyne.NewSize(400, 600))

	// make a bunch of fake messages
	ch := channel{Name: "Testchan", Header: "Test-Channel"}
	//for i := range 203_000_000 { //38.5gb ram
	//for i := range 203_000_00 { //4.4gb ram
	//for i := range 203_000 { //167mb ram
	for i := range 2030 { //mobile
		//for i := range 1 { //102mb ram
		msgstring := fmt.Sprintf("message number %d", i)
		for i := range i % 4 {
			msgstring += fmt.Sprintf("\narbitrary newline %d", i)
		}

		var img fyne.Resource
		if i%20 == 0 {
			// random messages carry an image for testing
			img = theme.FyneLogo()
		}
		ch.Messages = append(ch.Messages, message{
			Date: time.Date(1995, 1, 1, 12, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Second),
			Poster: account{
				Name: fmt.Sprintf("Poster %d", i%5),
				Desc: getGodsWordsAsString(12),
			},
			Message: msgstring,
			Image:   img,
		})
	}

	var content fyne.CanvasObject
	if !fyne.CurrentDevice().IsMobile() {
		content = container.NewAppTabs(
			container.NewTabItem("New Thing", g.newThing(&ch)),
			container.NewTabItem("Current Thing", g.howWeDidItBefore(&ch, 1000)), // 1000 is limit because it proves the point already
		)
	} else {
		// i only want that for the mobile testing
		content = g.newThing(&ch)
	}

	g.w.SetContent(content)
	g.w.ShowAndRun()
}

func (g *gui) newThing(ch *channel) fyne.CanvasObject {
	var makeObjects makeFuncInfiniteScroller
	if !fyne.CurrentDevice().IsMobile() {
		makeObjects = g.makeObjectsDesktop(ch)
	} else {
		makeObjects = g.makeObjectsMobile(ch)
	}

	is := newInfiniteScroller(func() int { return len(ch.Messages) }, makeObjects)
	is.GoToBottom()

	if !fyne.CurrentDevice().IsMobile() {
		g.w.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
			height := is.scroll.Size().Height * 0.7
			heightSmall := is.scroll.MinSize().Height
			switch ke.Name {
			case fyne.KeyPageDown:
				is.scroll.ScrollToOffset(is.scroll.Offset.AddXY(0, height))
			case fyne.KeyPageUp:
				is.scroll.ScrollToOffset(is.scroll.Offset.SubtractXY(0, height))
			case fyne.KeyDown:
				is.scroll.ScrollToOffset(is.scroll.Offset.AddXY(0, heightSmall))
			case fyne.KeyUp:
				is.scroll.ScrollToOffset(is.scroll.Offset.SubtractXY(0, heightSmall))
			}
		})
	}

	go func() {
		// constant stream of "messages" for testing
		var i int
		for t := range time.NewTicker(3 * time.Second).C {
			i++
			fyne.DoAndWait(func() {
				ch.Messages = append(ch.Messages, message{
					Date:    t,
					Poster:  account{Name: fmt.Sprintf("Tikka %d", i), Desc: "this is a bot, it has no soul"},
					Message: fmt.Sprintf("el massago: %s %s %s", t.String(), t.String(), t.String()),
				})
				is.GoToBottomIfAtBottom()
			})
		}
	}()

	specific := widget.NewEntry()
	specific.OnSubmitted = func(s string) {
		asint, err := strconv.Atoi(s)
		if err != nil {
			return
		}
		is.GoToSpecific(asint)
	}
	specific.PlaceHolder = "Type the index you want to go to"

	return container.NewBorder(
		container.NewHBox(
			widget.NewIcon(theme.MailSendIcon()),
			widget.NewLabel(ch.Name),
			widget.NewSeparator(),
			widget.NewLabel(ch.Header),
			//
			widget.NewSeparator(),
			widget.NewButton("TOP", func() { is.GoToTop() }),
			widget.NewButton("BOTTOM", func() { is.GoToBottom() }),
		),
		specific, nil, nil,
		is.GetCanvasObject(),
	)
}

func (g *gui) makeObjectsDesktop(ch *channel) makeFuncInfiniteScroller {
	return func(lbound, ubound int) []fyne.CanvasObject {
		// you must ensure that lbound and ubound are valid
		objs := make([]fyne.CanvasObject, 0, (ubound-lbound)*2) // times 2 because we append 2 objects per iteration
		for _, msg := range ch.Messages[lbound:ubound] {
			var img *canvas.Image
			if msg.Image != nil {
				img = canvas.NewImageFromResource(msg.Image)
				img.SetMinSize(fyne.NewSquareSize(200))
				img.FillMode = canvas.ImageFillContain
			}

			poster := widget.NewHyperlink(msg.Poster.Name, nil)
			poster.TextStyle = fyne.TextStyle{Bold: true}
			poster.OnTapped = func() {
				dialog.ShowInformation(msg.Poster.Name, msg.Poster.Desc, g.w)
			}
			msgstring := widget.NewLabel(msg.Message)
			msgstring.Wrapping = fyne.TextWrapWord
			msgstring.Selectable = true

			date := widget.NewLabelWithStyle(msg.Date.Format("02.01.2006\n15:04:05"), fyne.TextAlignCenter, fyne.TextStyle{})
			date.Selectable = true

			deleteself := newTappableIcon(theme.DeleteIcon(), func() {
				dialog.ShowInformation("delete this?", "do you really want to delete this message stub", g.w)
			})
			options := newTappableIcon(theme.SettingsIcon(), func() {
				dialog.ShowInformation("do other stuff", "this should prompt you to do other stuff with the message stub", g.w)
			})
			righthbox := container.NewHBox(deleteself, options, widget.NewLabel("" /*little spacer to not have options behind the scrollbar*/))
			righthbox.Hide()

			var finalcontent fyne.CanvasObject
			if img != nil {
				finalcontent = container.NewVBox(img, msgstring)
			} else {
				finalcontent = msgstring
			}

			content := container.NewBorder(
				container.NewBorder(nil, nil,
					// left
					container.NewHBox(
						widget.NewIcon(theme.AccountIcon()),
						poster,
					),
					// right
					righthbox,
				),
				nil,
				date,
				nil,
				container.NewBorder(nil, nil, widget.NewSeparator(), nil, finalcontent),
			)

			objs = append(objs,
				// obj 1
				NewAnyHover(content,
					func(_ *desktop.MouseEvent) {
						righthbox.Show()
					},
					nil,
					func() {
						righthbox.Hide()
					}),
				// obj 2
				widget.NewSeparator(),
			)
		}
		return objs
	}
}

func (g *gui) makeObjectsMobile(ch *channel) makeFuncInfiniteScroller {
	return func(lbound, ubound int) []fyne.CanvasObject {
		// you must ensure that lbound and ubound are valid
		objs := make([]fyne.CanvasObject, 0, (ubound-lbound)*2) // times 2 because we append 2 objects per iteration
		for _, msg := range ch.Messages[lbound:ubound] {
			var img *canvas.Image
			if msg.Image != nil {
				img = canvas.NewImageFromResource(msg.Image)
				img.SetMinSize(fyne.NewSquareSize(200))
				img.FillMode = canvas.ImageFillContain
			}

			poster := widget.NewHyperlink(msg.Poster.Name, nil)
			poster.TextStyle = fyne.TextStyle{Bold: true}
			poster.OnTapped = func() {
				dialog.ShowInformation(msg.Poster.Name, msg.Poster.Desc, g.w)
			}
			msgstring := widget.NewLabel(msg.Message)
			msgstring.Wrapping = fyne.TextWrapWord
			msgstringalt := NewSecondaryTapperLayer(msgstring, func(_ *fyne.PointEvent) {
				dialog.ShowInformation("Edit", "stub add copy message, deleteself and options", g.w)
			})

			date := widget.NewLabelWithStyle(msg.Date.Format("02.01.2006\n15:04:05"), fyne.TextAlignCenter, fyne.TextStyle{})
			datealt := NewSecondaryTapperLayer(date, func(_ *fyne.PointEvent) {
				g.a.Clipboard().SetContent(date.Text)
			})

			var finalcontent fyne.CanvasObject
			if img != nil {
				finalcontent = container.NewVBox(img, msgstringalt)
			} else {
				finalcontent = msgstringalt
			}

			objs = append(objs,
				// obj 1
				container.NewBorder(
					poster,
					nil,
					datealt,
					nil,
					container.NewBorder(nil, nil, widget.NewSeparator(), nil, finalcontent),
				),
				// obj 2
				widget.NewSeparator(),
			)
		}
		return objs
	}
}

func (g *gui) howWeDidItBefore(ch *channel, maxobjs int) fyne.CanvasObject {
	content := container.NewVBox()

	for i, msg := range ch.Messages {
		if i >= maxobjs {
			// legacy suffers so we stop
			// we made our point
			break
		}
		msgstring := widget.NewLabel(msg.Message)
		msgstring.Wrapping = fyne.TextWrapWord
		content.Objects = append(content.Objects,
			container.NewBorder(
				container.NewHBox(
					widget.NewIcon(theme.AccountIcon()),
					widget.NewLabel(msg.Poster.Name),
				), nil,
				widget.NewLabel(msg.Date.Format("15:04:05")),
				nil,
				container.NewBorder(nil, nil, widget.NewSeparator(), nil, msgstring),
			),
			widget.NewSeparator(),
		)
	}
	content.Refresh()
	scroll := container.NewVScroll(content)

	return container.NewBorder(
		container.NewHBox(
			widget.NewIcon(theme.MailSendIcon()),
			widget.NewLabel(ch.Name),
			widget.NewSeparator(),
			widget.NewLabel(ch.Header),
		), nil, nil, nil,
		scroll,
	)
}

// Terry Davis
var corpus = []string{"'kay", "African", "Angel", "BBC", "BRB", "Bam", "Boo", "Burp", "CIA", "California", "Catastrophic Success", "China", "Church", "Cosmos", "Dad", "Dudly Doright", "FBI", "GarryKasparov", "Ghost", "Give me praise", "God", "God is not mocked", "God smack", "Greece", "Greek to me", "Han shot first", "Hasta", "Heaven", "Hicc up", "HolySpirit", "I'll ask nicely", "I'll be back", "I'll get right on it", "I'll let you know", "I'll think about it", "I'm God and you're not", "I'm God who the hell are you", "I'm beginning to wonder", "I'm bored", "I'm busy", "I'm done", "I'm feeling nice today", "I'm gonna smack someone", "I'm good you good", "I'm grieved", "I'm impressed", "I'm in suspense", "I'm not dead yet", "I'm not sure", "I'm off today", "I'm on a roll", "I'm the boss", "I'm thrilled", "I'm tired of this", "IMHO", "I am not amused", "I be like", "I can't believe it", "I could be wrong", "I could swear", "I didn't do it", "I didn't see that", "I don't care", "I donno", "I forgot", "I give up", "I got your back", "I had a crazy dream", "I hate when that happens", "I have an idea", "I just might", "I love this", "I love you", "I made it that way", "I pity the fool", "I planned that", "I quit", "I see nothing", "I veto that", "I was just thinking", "I was sleeping", "Icarus", "If had my druthers", "Is that so", "Is that your final answer", "Isn't that special", "It's nice being God", "It grieves me", "Ivy league", "Japan", "Jedi mind trick", "Jesus", "King Midas", "Knock you upside the head", "LOL", "Make America Great Again", "Mars", "Mission Accomplished", "Mom", "Moses", "NOT", "NeilDeGrasseTyson", "Trump", "Oh Hell No", "Oh really", "Okilydokily", "One finger salute", "Oy", "Pope", "Putin", "Pullin the dragons tail", "ROFLMAO", "Russia", "Shakespeare", "Shalom", "Shhh", "StephenHawking", "SupremerCourt", "Terry", "That's gonna leave a mark", "That's my favorite", "The good stuff", "This is confusing", "Varoom", "Vegas", "Venus", "Watch this", "What", "What I want", "What are you doing Dave", "WooHoo", "Wow", "Yawn", "Yes you are", "Yo", "You can count on that", "You da man", "You fix it", "You get what you pray for", "You know", "Zap", "Zzzzzzzz", "a flag on that play", "a likely story", "a screw loose", "abnormal", "absetively posilutely", "absolutely", "act", "adjusted for inflation", "adultery", "after a break", "ahh", "ahh thats much better", "air head", "and the award goes to", "and then what", "angel", "anger", "application", "are you deaf", "are you feeling lucky", "are you insane", "are you sure", "arent you clever", "arrogant", "as a matter of fact", "astounding", "astronomical", "astrophysics", "atheist", "atrocious", "au revoir", "awesome", "awful", "ba ha", "bad", "bad ol puddytat", "baffling", "bank", "basically", "basket case", "bastard", "battle", "be happy", "be quiet bird", "beam me up", "because I said so", "beep beep", "begs the question", "bickering", "big fish", "biggot", "birds", "bizarre", "blessing", "boink", "boss", "break some woopass on you", "bring it on", "bummer", "busybody", "but of course", "by the way", "bye", "can you hear me now", "car", "catastrophe", "caution", "chaos", "charged", "charity", "check this out", "cheerful", "chess", "chill", "chill out", "choose one", "chump change", "church", "class  class  shutup", "clever", "climate", "close your eyes", "come and get me", "comedy", "commanded", "completely", "computers", "conservative", "cosmetics", "could it be   Satan", "couldn't be better", "couldnt possibly", "courage", "cowardice", "cracks me up", "crash and burn", "crazy", "cursing", "dance", "dang it", "daunting", "dean scream", "debt", "delicious", "delightful", "depressing", "desert", "didn't I say that", "dignity", "do I have to", "do it", "do not disturb", "do over", "do you get a cookie", "do you have a problem", "do you know what time it is", "do you like it", "do you want another", "doh", "don't count on it", "don't even think about it", "don't have a cow", "don't mention it", "don't push it", "don't worry", "downer", "drama", "driving", "duck the shoe", "dude such a scoffer", "earnest", "economy", "eh", "ehh a wise guy", "ehheh that's all folks", "employee", "employer", "end", "endeared", "endeavor", "endure", "energy", "enough", "enough said", "envy", "epic fail", "et tu", "everything's a okay", "evolution", "exorbitant", "experts", "exports", "fabulous", "face palm", "failure is not an option", "failure to communicate", "fake", "fancy", "far out man", "fer sure", "fight", "figuratively", "food", "fool", "fortitude", "foul", "freak", "frown", "fun", "funny", "furious", "gambling", "game changer", "game over", "geek", "genius", "ghastly", "ghetto", "glam", "glorious", "gluttony", "go ahead make my day", "good", "gosh", "gross", "grumble", "guilty", "guppy", "ha", "handyman", "hang in there", "happy", "happy happy joy joy", "hard working", "harder than it looks", "hate", "have fun", "he be like", "heads I win tails you lose", "heathen", "hello", "here now", "hey Mikey he likes it", "hey thats right", "hi", "high five", "high mucky muck", "hilarious", "hippy", "hit", "ho ho ho", "hobnob", "hold on a minute", "holier than thou", "holy grail", "home", "homo", "honestly", "honesty", "hooah", "hope", "hopefully", "horrendous", "hot air", "hotel", "how's the weather", "how about", "how about that", "how about those yankees", "how bout it", "how come", "how could you", "how do I put this", "how goes it", "how hard could it be", "how high", "huh", "humility", "humongous", "hurts my head", "husband", "hypocrite", "ice cream", "if and only if", "if anything can go wrong", "illogical", "imports", "impossible", "in a galaxy far far away", "in a perfect world", "in other words", "in practice", "in theory", "incoming", "incredibly", "industrious", "ingrate", "insane", "ipod", "is it just me or", "it'd take a miracle", "it's hopeless", "it's my world", "it figures", "it gets better", "it was nothing", "jealousy", "job", "jobs", "joke", "joker", "joking", "joy", "joyful", "just between us", "just lovely", "kick back", "kludge", "later", "laziness", "left field", "let's roll", "let's see", "let me count the ways", "liberal", "lift", "lighten up", "like like", "listen buddy", "little buddy", "little fish", "look buddy", "look on the brightside", "look out", "love", "lulz", "lust", "lying", "make my day", "manufacturing", "maybe I didn't make it clear", "meek", "meh", "merry christmas", "middle class", "mine", "mission from God", "mocking", "money", "mundo stoked", "music", "my bad", "my precious", "na na", "nasty", "naughty", "nerd", "nevada", "never happy", "news to me", "no more", "no more tears", "no news is good news", "no way dude", "no you cant", "nope", "not", "not a chance in hell", "not good", "not in kansas anymore", "not in my wildest dreams", "not that theres anything wrong", "not the sharpest knife in the drawer", "not too shabby", "now that I think about it", "now you tell me", "nut job", "obviously", "off the record", "oh come on", "oh my", "oh no", "oh oh", "ohh thank you", "oil", "okay", "on occassion", "on the otherhand", "once upon a time", "one more time", "one small step", "oops", "ordinarily", "other", "ouch", "outrageous", "over the top", "overflow", "pardon the french", "patience", "peace", "perfect", "persistence", "pet", "petty", "phasors on stun", "pick me pick me", "piety", "place", "play", "poor", "population", "potentially", "pow", "praise", "praying", "pride", "programming", "prosperity", "pwned", "qed", "quit", "quit it", "quite", "radio", "really", "recipe", "refreshing", "relax", "repeat after me", "repent", "resume", "reverse engineer", "revolution", "rich", "ridiculous", "rip off", "rocket science", "rose colored glasses", "roses are red", "rubbish", "run away", "saber rattling", "sad", "scorning", "scum", "segway", "service sector", "services", "sess me", "sex", "shist", "shucks", "silly human", "sing", "skills", "sky", "sloth", "slumin", "smack some sense into you", "small talk", "smart", "smile", "smurfs", "snap out of it", "so he sess", "so let it be done", "so let it be written", "soap opera", "special case", "spending", "spirit", "spoiled brat", "sports", "spunky", "stoked", "straighten up", "strip", "study", "stuff", "stunning", "super computer", "surprise surprise", "take the day off", "take your pick", "talk to my lawyer", "tattle tale", "taxes", "test pilot", "thank you very much", "that's all folks", "that's for me to know", "that's much better", "that's no fun", "that's your opinion", "thats just wrong", "thats laughable", "thats right", "the", "the enquirer", "theft", "theres no place like home", "these cans are defective", "think you could do better", "this might end badly", "threads", "tiffanies", "to infinity and beyond", "tomorrow", "tree hugger", "try again", "uh huh", "umm", "umm the other answer", "umm what now", "unemployment", "unsung hero", "vengeance", "vengeful", "vermin", "vice", "virtue", "voodoo", "vote", "walking", "wanna bet", "wastoid", "watch it buddy", "wazz up with that", "we ve already got one", "well I never", "well golly", "well obviously", "whale", "what's it to you", "what's the plan", "what's up", "what a mess", "what a nightmare", "what do you expect", "what do you want", "what have you done for me lately", "what luck", "what part of God do you not understand", "what planet are you from", "what the heck", "what was I thinking", "what would Jesus do", "whatcha talkin' 'bout", "whazza matter for you", "when hell freezes over", "where's the love", "whiner", "white trash", "who's to say", "who are you to judge", "whoop there it is", "why didn' you tell me", "why do I put up with this", "why is it", "wishful thinking", "won't you be my neighbor", "wonderbread", "wonderful", "woot", "wot", "wrath", "yada yada yada", "yeah", "yep", "yikes", "you'll see", "you're fired", "you're in big trouble", "you're lucky", "you're no fun", "you're not all there are you", "you're nuts", "you're out of your mind", "you're so screwed", "you're welcome", "you're wonderful", "you are my sunshine", "you better not", "you do it", "you don't like it", "you don't say", "you hoser", "you know a better God", "you never know", "you owe me", "you see the light", "you should be so lucky", "you shouldn't have", "you talkin' to me", "you think I'm joking", "you think you could do better", "yuck", "zoot", "red fang", "rum bitty di", "I m prettier than this man", "This cant be william wallace", "got the life", "king nun", "king of mars", "an Irishman is forced to talk to God", "you couldnt navigate yer way circleK", "its trivial obviously", "rufus!"}
var corpuslen = len(corpus)

func getGodsWordsAsString(count int) string {
	var resp strings.Builder

	resp.WriteString("God says:")
	for range count {
		resp.WriteRune(' ')
		resp.WriteString(corpus[rand.Intn(corpuslen)])
	}

	return resp.String()
}
