FROM alpine
RUN mkdir otf-classifier
COPY ./otf-classifier /otf-classifier
WORKDIR /otf-classifier/build/Linux64/lpofai_classifier
CMD [ "./lpofai_classifier" ]