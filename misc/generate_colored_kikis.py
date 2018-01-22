import os
import randomcolor

rand_color = randomcolor.RandomColor()

hues = ['red','orange','green','yellow','blue','purple','pink','monochrome']

kiki_num = 0
for i in range(1,9,1):
	for color in rand_color.generate(luminosity='bright',count=5):
		os.system("convert all/kikiset-0{}.png -fuzz 50% -fill '{}' -opaque red processed/kiki_{}.png".format(i,color,kiki_num))
		kiki_num += 1
	for hue in hues:
		for color in rand_color.generate(hue=hue,count=5):
			os.system("convert all/kikiset-0{}.png -fuzz 50% -fill '{}' -opaque red processed/kiki_{}.png".format(i,color,kiki_num))
			kiki_num += 1
