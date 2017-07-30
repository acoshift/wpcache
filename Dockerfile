FROM acoshift/go-scratch

USER 65534:65534
COPY wpcache /
ENTRYPOINT ["/wpcache"]
