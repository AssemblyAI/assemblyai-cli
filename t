import requests
import json

url = "https://api.assemblyai.com/v2/transcript"

file_path = "path/to/audio/file.mp3"

headers = {
    "authorization": "YOUR_API_KEY",
    "content-type": "audio/mpeg"
}

data = {
    "language_model": "xx-XX" # kde xx-XX reprezentuje kód jazyka
}

response = requests.post(url, headers=headers, data=open(file_path, "rb"), json=data)

print(response.json())
