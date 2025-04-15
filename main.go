package main

import (
	"time"

	"github.com/GolangProject/PixivCrawler/common/crawler"
)

func main() {
	cookie := "first_visit_datetime_pc=2025-03-27%2001%3A44%3A16; p_ab_id=6; p_ab_id_2=7; p_ab_d_id=652762805; yuid_b=ICgXlkY; device_token=bfec89d6f7b1367f4388d10a6d7aa088; privacy_policy_notification=0; a_type=0; b_type=1; _gid=GA1.2.839493238.1744419787; _gcl_au=1.1.699141397.1744423676; login_ever=yes; __utmz=235335808.1744534718.10.6.utmcsr=zhidao.baidu.com|utmccn=(referral)|utmcmd=referral|utmcct=/question/628727835479498332.html; __utmc=235335808; __utma=235335808.2020838566.1743609792.1744534718.1744540597.11; cf_clearance=m3KnE5ovz4RFg91G42SV2hJ6bFtkgP7FZAaohRFAvVs-1744547809-1.2.1.1-FwpmiJNaQVDmfHkBERww88tvxpkLWjQU0gJPQGmfQdozxxyPMuLHvnEFwNzyMIivjMNGE6hydI.iiTaCpAQpvMinQlCuDYNMcydxHL6wDf7_frQTNYeOQQ8tWMGYSnjzkPhW8O_ce0PR74LGschHpEdKQwix3i7w3Zsh5h2yCXSbfrpLrH26Tu9o2wLZGz9QL883w5reAInaGb7pIOZ.HqW9mtJTJPtitCXgQj3BLWxOPeTNd5s4AS5bfX.I.3p2w6n_Tib33.HNqLsLk2jcHoKyRVyvnpJLcrbA1ycIreKnTczD0hrShK_lBxipBQr7AQ0DRkFpQuCV7Pwc9znGbURmZMLbkAD66Xx9WYWqhF8; cc1=2025-04-13%2021%3A37%3A41; PHPSESSID=25368779_4mRvutN010lcWWk353uCdQKwGdt8xwKX; _ga_MZ1NL4PHH0=GS1.1.1744548417.4.0.1744548423.0.0.0; c_type=22; privacy_policy_agreement=0; __utmv=235335808.|2=login%20ever=yes=1^3=plan=normal=1^5=gender=male=1^9=p_ab_id=6=1^10=p_ab_id_2=7=1; __utmt=1; __cf_bm=Tu.PDT2QwV9tA8PwlTl9r3Vy1hSFKAvrcopaja99pbI-1744548428-1.0.1.1-hKh6jE1_XxHPUbhkKaHODzzi859_ShzuqvGPfsU8Z8_CFA_pNxhexHzoqrB_NWVAGgLcLXJ2ec9QrZwKQQGQIHN3ADu89i.vLFlxAYLstPmRtZedY0PDISQoMUrNY9a4; _ga=GA1.2.2020838566.1743609792; _gat_UA-1830249-3=1; FCNEC=%5B%5B%22AKsRol9iL35fGQJs_QgfY3MLxblhcOiOznGpJJf2DTpw8TOgA-bWEXh5lGqm1ySLkDxg2jr63sSbiw3eWEx6MnJlhRkV8flgH-FtUwTD5rAUsEiYtFVOp_mspnwlnBjrub7uWNzJR8ThuBTtPmlPtnLZYJYluyuDIQ%3D%3D%22%5D%5D; __utmb=235335808.20.10.1744540597; _ga_75BBYNYN9J=GS1.1.1744547784.12.1.1744548467.0.0.0"
	agent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0"
	accept := "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"

	config := crawler.InitUserConfig(0)
	config.SetCookie(cookie)
	config.SetAgent(agent)
	config.SetAccept(accept)
	config.SetDelay(2 * time.Second)

	crawler := crawler.NewCrawler(config)
	crawler.Run()
}
