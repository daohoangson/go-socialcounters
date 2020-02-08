FROM google/cloud-sdk

RUN apt-get update

RUN apt-get install -y golang

WORKDIR /src

CMD ["bash"]
