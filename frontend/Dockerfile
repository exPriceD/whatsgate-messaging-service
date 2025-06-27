FROM nginx:alpine

# Копируем index.html, app.js, pages/ и ui/
COPY . /usr/share/nginx/html/

# Заменяем основной nginx.conf (а не кладём в conf.d)
COPY nginx.conf /etc/nginx/nginx.conf
