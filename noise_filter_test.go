package hotspider

import "testing"

func TestDefaultNoiseFilter(t *testing.T) {
	filter := DefaultNoiseFilter()
	noise := []string{
		"身为魔王的我娶了精灵为妻该如何表白我的爱",
		"十八岁太奶奶驾到，重整家族荣耀",
		"穿越荒年，刁蛮老太重掌全家",
		"吞下人鱼珠我成了商界大亨",
		"我真没想重生啊·动态漫",
		"宝宝巴士儿歌3D版",
	}
	normal := []string{
		"余生，请多指教",
		"你好，李焕英",
		"知否知否，应是绿肥红瘦",
		"我才不要和你做朋友呢",
		"新大头儿子和小头爸爸之天眼系列失控的第七天",
		"画江湖之不良人第七季：九幽玄天",
		"我们恋爱吧 第六季",
	}

	for _, title := range noise {
		if !filter.IsNoise(title) {
			t.Errorf("IsNoise(%q) = false, want true", title)
		}
	}
	for _, title := range normal {
		if filter.IsNoise(title) {
			t.Errorf("IsNoise(%q) = true, want false", title)
		}
	}
}

func TestCustomNoiseFilter(t *testing.T) {
	filter := &NoiseFilter{
		Keywords: []string{"奥特曼"},
		Custom: func(title string) bool {
			return title == "空枪"
		},
	}

	tests := map[string]bool{
		"迪迦奥特曼[普通话版]": true,
		"空枪":          true,
		"身为魔王的我娶了精灵为妻该如何表白我的爱": false,
		"宝宝巴士儿歌3D版":            false,
	}
	for title, want := range tests {
		if got := filter.IsNoise(title); got != want {
			t.Errorf("IsNoise(%q) = %v, want %v", title, got, want)
		}
	}
}

func TestSetNoiseFilter(t *testing.T) {
	defer SetNoiseFilter(DefaultNoiseFilter())

	const title = "宝宝巴士儿歌3D版"
	if !isNoiseTitle(title) {
		t.Fatalf("default filter should match %q", title)
	}

	SetNoiseFilter(nil)
	if CurrentNoiseFilter() != nil || isNoiseTitle(title) {
		t.Fatal("nil filter should disable filtering")
	}

	SetNoiseFilter(&NoiseFilter{Keywords: []string{"3D版"}})
	if !isNoiseTitle(title) {
		t.Fatalf("custom filter should match %q", title)
	}
}
