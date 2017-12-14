serve:
	hugo server --ignoreCache -D --watch -t twotwothree -b http://localhost:1313 --bind 0.0.0.0 

production:
	hugo server --appendPort=false --baseURL https://projectx.docs/ --bind 127.0.0.1 --port 8131 --disableFastRender --ignoreCache --watch --theme twotwothree

build:
	hugo -t twotwothree
