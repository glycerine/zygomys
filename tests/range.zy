;; range loop components: mdef and hpair

;; mdef
(mdef a b c (list 4 5 6))
(assert (== a 4))
(assert (== b 5))
(assert (== c 6))

;; hpair
(def h (hash a:3 b:9 c:27 d:-4))
(assert (== (first (hpair h 0)) a:))
(assert (== (first (hpair h 1)) b:))
(assert (== (first (hpair h 2)) c:))
(assert (== (first (hpair h 3)) d:))

(assert (== (second (hpair h 0)) 3))
(assert (== (second (hpair h 1)) 9))
(assert (== (second (hpair h 2)) 27))
(assert (== (second (hpair h 3)) -4))

(defmac range [key value my-hash & body]
  ^(let [n (len ~my-hash)]
      (for [(def i 0) (< i n) (def i (+ i 1))]
        (begin
          (mdef (quote ~key) (quote ~value) (hpair ~my-hash i))
          ~@body))))


;; verify that range over hashes works
(def h (hash a:44 b:55 c:77 d:99))
(def s "")
(range k v h (set s (concat s " " (str k) "-maps->" (str v))))
(assert (== s " a-maps->44 b-maps->55 c-maps->77 d-maps->99"))

