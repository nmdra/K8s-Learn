The **`ctr`** command-line tool is a lightweight CLI provided by **containerd** to interact directly with its API.

---

https://labs.iximiuz.com/challenges/copying-files-from-container-images-with-ctr

```bash
   ctr image pull registry.iximiuz.com/tricky-one:latest
   ctr run --rm -t registry.iximiuz.com/tricky-one:latest tricky-instance sh
   find / -name tricky-file.txt
   ctr snapshot mount <snapshot-path> <target-path>  # Inspect and copy manually
```
