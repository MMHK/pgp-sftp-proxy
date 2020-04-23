package lib

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func Test_PGP_Encrypt(t *testing.T) {
	sourceFile, err := ioutil.ReadFile(getLocalPath("../test/isrd_id_card_path2018061314121908487.jpg"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	//
	//keyFile, err := ioutil.ReadFile(getLocalPath("../test/test-key.pem"))
	//if err != nil {
	//	t.Log(err)
	//	t.Fail()
	//	return
	//}
	reader  := strings.NewReader(`-----BEGIN PGP PUBLIC KEY BLOCK-----  mQENBF6ZKp0BCADRhlwr8oVOtmko0eI7aYWCKldp3ahtgEzQtSHa9ax2t3hheoT8 cSKsp3lWBge0F7p4l4Jxu45qFOJAI3eqzmKIMaTElYpGhRKjb0NHBB3iMoe2uWdh AlHZmW/VW8T71en4Mjd5z+dZViczzrurwxe9SrZOSzPUI+T6Y7AR+P4SxeSE1cXr tgiOSksdGdEjjiKnkdV5O8k8AtDrE9Zt65+TCHFmVh8QcO7RHBjz3zXUlQyissIX W75+j451rxiPUJ8px5U4JBZcu4VJJOWfVRpx6IAk7sJZfz3I5IT/O93RNf6BlWh0 B2mX+7PyKVSzB9G+D/m/VDLNWQpwLc89N0ZRABEBAAG0G1NhbS1UZXN0IDxzYW1A bWl4bWVkaWEuY29tPokBVAQTAQgAPhYhBDWrXiUG6xtZc/kSb6GFjeyZLrZlBQJe mSqdAhsDBQkDwmYjBQsJCAcCBhUKCQgLAgQWAgMBAh4BAheAAAoJEKGFjeyZLrZl +skIAKAdnA1LuNWYohEW0SrniMUjidM59HGzzRB6UzESQbnm0+l44gm/jG+Suf3x 5tAcoVOC+Qd+CtWmKr4hFwsasTW4Y59nesn9iBUovvDW7j5mELovxMFIsZ5brUg/ RZEFnwOChdhKFCDp3beQWfSl3zTuLSu77U+XZvcD0ceV1twSQtqlOusPKYoE4LEN 91CNbgQK9V8I5h2gW06CPwOGoy0+FbnKvZYqYmNrmhe7qFkRq2kK4qyfLX0vvDUA 71CVqab1pVgCfylti1l/07Tlm0o8zKieFnxPpmwRgOKtHp9iDTG5ftMsJLhD7IIv /yO7qhS00WPescGKJ8Nsu+9Ully5AQ0EXpkqnQEIAL9KBx4LQPBo1uFsui8k4277 CEnw/4gp0oQ0xr4mrNseImI7U6Q7LC61MI9NrYc2Ea0xjP/QqSwadrLFh64IEPig hA1Pi6TmDpLvEbp0k5SHjYpMl8thJ5LB6u3ONgGiDLnlhc2laDTDVizQ0knIusjC zs9O0DyNEZ0db5eajUGC+yYhOh+Z9VzViXkbd9JcRNzw93gAjuuDvcG1kQHSIZeo X7pRH8e6untovbPdtHWMJ2HkAHcti/P62dBJ5rIrKPt1gG4/4yAjE/nwDNO+edGt sNjCj3SzaRdvg0MZgdiX5F7Ws3M006C1G/O8VDJ+bYcxiCI09fswtWB+K07a0EkA EQEAAYkBPAQYAQgAJhYhBDWrXiUG6xtZc/kSb6GFjeyZLrZlBQJemSqdAhsMBQkD wmYjAAoJEKGFjeyZLrZlEkEH/0FLj1/7Aq2Mcibkdl5s/uLyeHdCTWV7twuifCfY oaSwal1AhOwsILq3ibEI6ukYC1ggu+ns+xJBUdgZJPzLuRuLj+7xS5WqZZJEIIL6 4IMq/87rEQ2aEodsqdHDPA9TjNpo/ceCGsLW2BcsQfeMA29ApAfqnJ9S7SagLHkT b9cgNVy4eqrHUU10atmu1KLv0uuT8VqLYNtRyCqCPPEy7dBfPFbdZTpdvw+Pj6WG /v5BoP/3Ug9ry3TObveait2sujoPQcO1Fz7fL8CYmdrGGx1sRFA38XOYco7zlOfa ZHf2Rts5hd2X++1+A4FbUe6dHofdnkE+txkCrKpAtyHdiWM= =McYr -----END PGP PUBLIC KEY BLOCK-----`)


	pgp, err := PGP_Encrypt(sourceFile, reader)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(pgp)
}

func Test_PGP_Encrypt_File(t *testing.T) {
	sourceFile, err := ioutil.ReadFile(getLocalPath("../test/isrd_id_card_path2018061314121908487.jpg"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	keyFile, err := os.Open(getLocalPath("../test/test-key.pem"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	distFile := getLocalPath("../test/isrd_id_card_path2018061314121908487.jpg.pgp")

	err = PGP_Encrypt_File(sourceFile, keyFile, distFile)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
}
