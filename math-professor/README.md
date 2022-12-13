To upload,

```
  zip myfn.zip *.py
  aws update-function-code --function-name mathFunction --zip-file fileb://$(pwd)/myfn.zip
```
