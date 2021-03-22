# goFuzz

**How to use goFuzz**
1. Put goFuzz under the same GOPATH of the target application
    
2. Use goFuzz/runtime to overwrite the original runtime 
   
   Note: goFuzz/runtime is based on the runtime of go-1.14.2
   
   Remember to have a backup of the original runtime.
       
3. Use goFuzz/cmd/instrument to insert necessary function 
calls into the target application

    TODO: this step hasn't been not completed yet