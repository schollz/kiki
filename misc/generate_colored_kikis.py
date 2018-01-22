import os
import randomcolor

rand_color = randomcolor.RandomColor()

for i,color in enumerate(rand_color.generate(luminosity='bright',count=100)):
	os.system("convert kiki.png -fuzz 50% -fill '{}' -opaque red kiki_{}.png".format(color,i))
