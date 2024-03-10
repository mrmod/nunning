FROM alpine:latest

WORKDIR /app
COPY ../_dist/release /app/homewatch
COPY ../_dist/entrypoint.sh /app

# Expose prometheus metrics port
EXPOSE 2112
ENTRYPOINT ["/app/entrypoint.sh"]