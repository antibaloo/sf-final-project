Итоговый проект курса (PJ-04) студента потока GO-37

Проект реализует 4 сервиса:

Сервис новостей (news) - запускается по localhost:8081
В составе сервиса следующие обработчики:
- метод вывода списка новостей: /GET /news?search=...&page=.&request_id=xxxxxxx
    Возвращает json структуру страницы с номером, переданном в параметре page, или первую, если параметр отсутствует, списка новостей, заголовки которых содержать слово переданное в параметре search (необязательный), и структуру объекта паджинации,
    содержащий: количество новостей на страницуб номер страницы, количество страниц. 

- метод получения детальной новости GET /news/{id}/detailed?request_id=xxxxxxx
    Возвращает json структуру со всеми полями новости с заданным идентификатором

Во все запросы сервиса новостей шлюз передает параметр request_id - индентификатор запроса, используется при логировании.

Сервис комментариев (comments) - запускается по localhost:8082
В составе сервиса следующие обоработчики:
- метод добавления комментария к новости: /POST /comment
    Добавляет комментарий в БД
- метод получения все комментариев к конкретной новости: GET /news/{id}/comments
    Возвращает список всех комментариев к новости с переданным идентификатором

Во все запросы сервиса комментариев шлюз передает параметр request_id - индентификатор запроса, используется при логировании.

Сервис новостей меет в своем составе метод чтения новостей из rss канала, который запускается в отдельной горутине для каждого канала, читает из него новости по таймауту и записывает их в БД

Сервис проверки комментариев - запускается по localhost:8081
При запуске сервис читает из БД список запрещенных слов, а едиственный обработчик проверки комментария POST /check
    проверяет нет ли в переданном комментарии слов из упомянутого списка, в случае успешного контроля возвращает 200, в противном случае 400.

Во все запросы сервиса проверки комментариев шлюз передает параметр request_id - индентификатор запроса, используется при логировании.

Сервис шлюза (API Gateway) - запускается по адресу localhost:8080
В составе сервиса следующие обработчики:
- метод вывода списка новостей: GET /news
    Метот отправляет запрос к сервису новостей и возвращает клиенту список новостей в сооттветствии с заданными параметрами или ошибку.
- метод вывода детальной новости: GET /news/{id}
    Метод асинхронно отправляет запрос к сервису новостей, чтобы получить тектст конкретной новости и запрос к сервису комментариев, чтобы получить список комментариев к конкретной новости и возвращает клиенту структуру детальной новости
- метод добавления комментариев: POST /ceomment
    Метод отправляет запрос к сервису проверки комментарием и, если проверка было пройдена, запрос к сервису комментариев, возвращая клиенту результат операции.

Если от клиента поступил параметр request_id, он переправляется внутренним сервисам, если такого параметра нет, идентификатор генерится сервисом и передается к внутенним сервисам.

Все сервисы системы ведут логирование зпросов и ответов обработчиков.