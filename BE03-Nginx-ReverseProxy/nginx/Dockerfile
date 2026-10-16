FROM nginx:1.27-alpine

# Copy custom Nginx configuration
COPY nginx/nginx.conf /etc/nginx/nginx.conf

# Copy static website files (index.html, image.png)
COPY static/index.html /usr/share/nginx/html/index.html
COPY static/image.png /usr/share/nginx/html/image.png

# Copy SSL certificates
COPY ssl/server.crt /etc/nginx/ssl/server.crt
COPY ssl/server.key /etc/nginx/ssl/server.key

# Prepare cache and run directories
RUN mkdir -p /var/cache/nginx && chown -R nginx:nginx /var/cache/nginx /usr/share/nginx/html

EXPOSE 80 443
CMD ["nginx", "-g", "daemon off;"]
