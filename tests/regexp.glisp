(let* 
     [re (regexp-compile "hello")
      loc (regexp-find-index re "012345hello!")]
  (assert (== (aget loc 0) 6))
  (assert (== (aget loc 1) 11))
  (assert (== "hello" (regexp-find re "ahellob")))
  (assert (regexp-match re "hello"))
  (assert (not (regexp-match re "hell"))))
