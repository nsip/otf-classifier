echo '\ncurl -i -X GET http://localhost:1576/align/\n\n'
curl -i -X GET http://localhost:1576/align
echo '\ncurl -i -X GET http://localhost:1576/align?text=information/\n\n'
curl -i -X GET http://localhost:1576/align?text=information
echo '\ncurl -i -X GET http://localhost:1576/align?area=Numeracy&text=information\n\n'
curl -i -X GET "http://localhost:1576/align?area=Numeracy&text=information"
echo '\ncurl -i -X GET http://localhost:1576/align?area=Numeracy&text=collects%20information\n\n'
curl -i -X GET "http://localhost:1576/align?area=Numeracy&text=collects%20information"
echo '\ncurl -i -X GET http://localhost:1576/index?search=understanding\n\n'
curl -i -X GET http://localhost:1576/index?search=understanding
