How to ensure new versions are on Github and Homebrew

When files are updated, create a new tag 


```git tag v0.1.{}   
   git push origin v0.1.{}
```

Run a checksum on the tag

```
curl -sL "https://github.com/hanif-adedotun/mirusync/archive/refs/tags/v0.1.{}.tar.gz" | shasum -a 256
```

Copy the checksum to @homebrew-mirusync/Formula/mirusync.rb


Then run 
```
brew tap hanif-adedotun/mirusync
brew update
brew install mirusync
```