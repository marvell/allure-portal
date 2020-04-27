# AllurePortal #

## Local usage ##

To try you can use docker:

```bash
docker run --name=allure-portal -d -p 80:80 -v $(pwd)/storage:/storage marvell/allure-portal --base-url=http://localhost
```

Then try to upload ZIP archive with allure results (json files), don't forget to change placeholders:

```bash
curl -F file=@<zip-file> "http://localhost/upload?group=<group>&project=<project>&version=<version>"
```

Example
```bash
curl -F file=@resources/allure_results_0.zip "http://localhost/upload?group=test&project=review&version=0"
```
