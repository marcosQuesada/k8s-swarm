FROM golang:1.17.5 as builder

ARG DATE
ARG COMMIT
ARG SERVICE

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN echo "building k8s swarm service: ${SERVICE} on commit: ${COMMIT} date: ${DATE}"

RUN CGO_ENABLED=0 go build -ldflags "-X github.com/marcosQuesada/k8s-swarm/config.Commit=${COMMIT} -X github.com/marcosQuesada/k8s-swarm/config.Date=${DATE}" ./services/${SERVICE}

# final stage
FROM alpine:3.11.5
ARG SERVICE
COPY --from=builder /app/${SERVICE} /app/
COPY --from=builder /app/services/${SERVICE}/config /app/config