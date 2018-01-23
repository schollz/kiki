import json, requests

content = """**My source code:**

```
import json, requests
quote_response = requests.get("https://quotes.schollz.com/subject/friend.json")
quote = quote_response.json()[0]
if len(quote['Name']) == 0:
	quote['Name'] = 'Unknown'
content = "<p>“{q[Text]}”</p><p>- <em>{q[Name]}</em></p>".format(q=quote)
payload={'content':content,'purpose':'share-text','to':['public']}
r = requests.post("http://localhost:8003/letter", data=json.dumps(payload))
print(r.json())
```
"""
payload={'content':content,'purpose':'share-text','to':['public']}
r = requests.post("http://localhost:8003/letter", data=json.dumps(payload))


quote_response = requests.get("https://quotes.schollz.com/subject/friend.json")
quote = quote_response.json()[0]
if len(quote['Name']) == 0:
	quote['Name'] = 'Unknown'
content = "<p>“{q[Text]}”</p><p>- <em>{q[Name]}</em></p>".format(q=quote)
payload={'content':content,'purpose':'share-text','to':['public']}
r = requests.post("http://localhost:8003/letter", data=json.dumps(payload))
print(r.json())

