import os
import randomcolor

rand_color = randomcolor.RandomColor()

hues = ['monochrome']

kiki_num = 0
for i in range(1,9):
	for hue in hues:
		os.system("convert all/kikiset-0{}.png -fuzz 50% -fill '{}' -opaque red processed/kiki_{}.png".format(i,'#b7b7b7',kiki_num))
		kiki_num += 1
