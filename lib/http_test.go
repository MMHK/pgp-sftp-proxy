package lib

import (
	"bytes"
	"encoding/json"
	"testing"
)

func Test_DecodeBody(t *testing.T) {
	source := `{"files":{"M_Article_Zurich_2.gif":"https:\/\/s3-ap-southeast-1.amazonaws.com\/s3.jetso.com\/asset\/images\/M_Article_Zurich_2.gif","M_Article_Zurich_4.jpg":"https:\/\/s3-ap-southeast-1.amazonaws.com\/s3.jetso.com\/asset\/images\/M_Article_Zurich_4.jpg","M_Article_Zurich_ca.gif":"https:\/\/s3-ap-southeast-1.amazonaws.com\/s3.jetso.com\/asset\/images\/M_Article_Zurich_ca.gif"},"key":"-----BEGIN PGP PUBLIC KEY BLOCK-----\r\nVersion: GnuPG v2\r\n\r\nmQENBFV2aQkBCADuEi0WB\/VeHp2zo\/6XRnX6uLbIyKQszo0gW6Ek4WGdTvovX\/9r\r\nh6qNx++pcLmT8wmuwvMMIyvsNEt5eKsWSgjJZfSqwo2uMYePpz2ZjruC+eGzONS5\r\nnWBbmScnmGphlLXnW8OpOb2JFqiZRj8Rv+UEUy39DsFiwsNBRkYzWgbX6yI7YgNH\r\nRZxcCWvhZZrDbBSDlhzzSFQttVS+PchvI1rXkgbO5igopsolj86LnB0HnZqlivNE\r\naQ1xxfTKPv9tKm3DeZqEPdbpBkxBdrqDEye9Gjq06wgJQ68bIzwqAAFuCKKWfeCg\r\nCclw3MaVTXX5wuwl4V8mqVvkMOUt9Qkli149ABEBAAG0J0RSSVZFUl9VQVRfS0VZ\r\nIDx0ZXJyYW5jZUBkcml2ZXIuY29tLmhrPokBOQQTAQgAIwUCVXZpCQIbDwcLCQgH\r\nAwIBBhUIAgkKCwQWAgMBAh4BAheAAAoJENxsjOiA47hHuxIH\/3Y8DgLiM0oD6opP\r\nN1Wwnd5f9\/J3is9WlaKuxGP6iDjHKfTf2Bcwm5AC1+XosW6HSrd7g9JiubG6Cvsz\r\nkI\/voFVGPJoCr+2sPY0r8hCYFQAYPr1U9EoCTYICORbJZMeucAo4v6AH9LxwDFx0\r\n8IpXkfwett+Q2AvMAQw6v0s0bqTJ20n4dLCfhdu3IdDgTXlg6My\/mGswao1f+BdE\r\ntdJ5iBL\/QMpowoz2SZeiYMtLOxf+NC5h2iVxd+ijZ0JjMEedSozz0y60QuVWaJ2J\r\nndSjEwhphcx6cGctnJ83w4CQkGurfGQKs0S5k+5zUxANulSufSiH9mC4n3rOEw2v\r\nqW3H6jg=\r\n=YW+U\r\n-----END PGP PUBLIC KEY BLOCK-----\r\n","env":"dev","notify":"http:\/\/www.baidu.com\/"}`
	buffer := bytes.NewBufferString(source)
	decoder := json.NewDecoder(buffer)
	var reqBody MultipleBody
	err := decoder.Decode(&reqBody)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	t.Log(reqBody)
	t.Log("PASS")
}