FROM golang:alpine AS builder

COPY . /app/
WORKDIR /app
RUN CGO_ENABLED=0 go build -v -trimpath -ldflags "-s -w" -o ssws .

FROM scratch
COPY --from=builder --chown=0:0 /app/ssws /ssws
CMD [ "/ssws" ]
