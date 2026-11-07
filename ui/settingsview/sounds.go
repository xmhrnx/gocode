package settingsview

import "gocode/ui/core"

func soundEntries(s *core.Services) []entry {
	so := &s.Settings.Sounds
	return []entry{
		{title: "Play sounds", description: "Short, soft tones when something needs you. Turn this off for silence.",
			keywords: "sound audio chime beep notification mute silent", control: switchControl(s, &so.Enabled)},
		{title: "When the agent is done", description: "Two rising notes when a run finishes.",
			keywords: "sound done finished complete chime", control: switchControl(s, &so.Done)},
		{title: "When the agent asks for permission", description: "Two even notes when a tool call waits for your answer.",
			keywords: "sound permission approval ask attention", control: switchControl(s, &so.Permission)},
		{title: "When a run fails", description: "Two falling notes when a run stops with an error.",
			keywords: "sound error failure", control: switchControl(s, &so.Errors)},
		{title: "Volume (%)", description: "How loud the sounds are.",
			keywords: "sound volume loud quiet", control: number(s, &so.Volume, 0, 100)},
		{title: "Listen", description: "Play each sound at the volume above.",
			keywords: "sound preview test listen", control: newSoundPreview(s)},
	}
}
