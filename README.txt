HTTP-сервис отвечает на запросы в которых содержится uid,
после чего в возвращает/создает соотвествующий документ MongoDB
для уменьшения обращений к Mongo используется memcached

В случае успеха возвращает 200 и (не обязательно) JSON
В случае любой ошибки или если данные не найдены возвращает 503
(TODO: возвращать 404 если нет данных)

Хендлеры

* /version (GET) - ищет по uid документ, если находит – обновляет кеш и возвращает JSON
    {version: версия}

* /data (GET) - ищет по uid документ, если находит и передан параметр counter
    возвращает поле tags документа, если counter не задан – возвращает data

* /set (POST) -  обновляет переданные в запросе поля документа с заданным в запросе uid
или создает его, если его еще нет (поле version инкрементируется на один)
обновляет кеш
возвращает только HTTP-код

* /setcounter (POST) - если передан не пустой параметр tg, разбивает значение переданные в нем
строку, через точку. Полученные значения используются как ключи в поле tags документа (по uid)
значение переданных ключей увеличивается на еденицу
поле las_visit обновляется текущим таймстемпом


Подсказки:
uid - связанный с пользователем специальный ID