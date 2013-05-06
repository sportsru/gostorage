Краткое описание
================

HTTP-сервис отвечает на запросы в которых содержится uid (выдается nginx-ом),
после чего в возвращает/создает соответствующий документ MongoDB в зависимости от
обработчика. Для уменьшения обращений к сервису, сохраняет версию ответа в мемкеше
(используется внешним прокси, nginx-ом).

Хендлеры:
---------

* /version (GET) - ищет по uid документ, если находит – возвращает JSON вида    {version: версия}, если не находит то возвращает версию -1, Обновляет кеш
* /data (GET) - ищет по uid документ, если находит и передан параметр counter
    возвращает поля tags документа в формате JSON, если counter не задан – возвращает поля data, в случае если документ не найден возвращает пустой json
* /set (POST) -  обновляет переданные в запросе поля data документа с заданным в запросе uid или создает его, если его еще нет (поле version инкрементируется на один); возвращает только HTTP-код
* /setcounter (POST) - если передан не пустой параметр tg, разбивает значения и пишет в поля tags ключи с инкрементацией значений этих ключей на 1

поле last_visit обновляется текущим таймстемпом в секундах


Пример конфига nginx:
-------------------

	upstream storage_cluster {
        server 192.168.1.240:3000 weight=1 max_fails=3 fail_timeout=10s;
        server 192.168.1.65:9002 weight=5 max_fails=3 fail_timeout=10s;
        server 192.168.1.66:9002 weight=5 max_fails=3 fail_timeout=10s;
        keepalive 32;
	}

	location @user_crossdomain_fallback {
        rewrite ^/crossdomain/config/(.*) /$1 break;
        proxy_send_timeout 10s;
        proxy_pass      http://storage_cluster;
	}

	location /crossdomain/config/version/ {
        set                             $memcached_key          "s_$arg_uid";
        memcached_pass                  192.168.1.240:11211;
        memcached_next_upstream         not_found;
        add_header      Cache-Control   "max-age=0, must-revalidate";
        error_page      404 502 504 = @user_crossdomain_fallback;
	}

	location /crossdomain/config/data/ {
        proxy_cache_use_stale updating http_502;
        proxy_cache site;
        proxy_cache_lock on;
        proxy_cache_key $request;
        proxy_cache_valid 120m;
        proxy_send_timeout 5s;
        proxy_pass      http://storage_cluster;
	}

	location /crossdomain/config/set {
        rewrite ^/crossdomain/config/(.*) /$1 break;
        proxy_send_timeout 5s;
        proxy_pass      http://storage_cluster;
	}



