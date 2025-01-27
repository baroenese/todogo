# Sumber Inspirasi

Proyek ini terinspirasi dari chanel youtube [Review Kodingan Go Yang Sangat Tidak Clean! #Codejam Imre Nagi](https://www.youtube.com/live/PxlHx6Whcic?si=6yQCrbvPjBXA9wO8)

Ucapan terima kasih yang sebesar-besarnya kepada Mas Didit Koding aja dulu [KJM](https://www.youtube.com/@KodingAjaDulu) atas ilmu yang telah dibagikan🙏🏽

Source asli dapat dilihat pada [Module Driven Go Todolist](https://github.com/kadcom/mda)

## Database

Konfigurasi database dapat dilihat pada repositori __todogo-migration__

# Modul-Driven Go

Berbeda dengan yang disebut sebagai *'standard go structure'*, proyek ini bertujuan untuk menunjukkan bahwa untuk membuat perangkat lunak yang mengikuti arsitektur bawang *(onion architecture)* di Go, Anda tidak perlu membuat banyak objek dan struktur.

## Motivasi

Ketika saya melihat proyek-proyek di luar sana, atau ketika saya membaca proyek dari klien saya, ada pola yang saya kenali. Kebanyakan dari mereka mencoba menyesuaikan Go dengan model mental framework yang mereka kenal.

Jadi, saya lelah melihat kode yang berbelit-belit yang mencoba MEMAKSAKAN orientasi objek ke Go. Inilah mengapa proyek ini menggunakan __*mostly functions*__.

## Prinsip Membangun Proyek Ini

Proyek ini bukan contoh sempurna dari perangkat lunak tingkat produksi. Saya membuat objek ini dengan prinsip-prinsip berikut:

- Mudah dipahami dan diajarkan.
- Idiomatik. Ini berarti saya menggunakan Go seperti yang dimaksudkan. Gunakan nilai sebanyak mungkin dan gunakan pola seperti *'accept interface returns value'* dan hindari hal-hal seperti mengembalikan antarmuka.
- *High cohesion, lose coupling.* Hal-hal yang seharusnya bersama harus berada di tempat yang sama.
- Dependensi yang masuk akal. Saya memiliki alasan mengapa saya menggunakan dependensi tersebut.
- Default yang masuk akal. Setiap konfigurasi memiliki default yang memudahkan pengembangan.

Proyek ini sesederhana vanilla. Saya tidak menggunakan ORM apa pun dan berusaha memiliki dependensi yang sangat minimal. Bahkan, proyek ini tidak memiliki antarmuka yang didefinisikan. Jika Anda melihat file `go.mod`, Anda akan melihat bahwa ini adalah dependensi yang saya gunakan:

```
require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/go-chi/chi/v5 v5.2.0
	github.com/jackc/pgx/v5 v5.7.2
	github.com/oklog/ulid/v2 v2.1.0
	github.com/rs/zerolog v1.33.0
	gopkg.in/guregu/null.v4 v4.0.0
	gopkg.in/yaml.v3 v3.0.1
)
```

Saya dapat menjelaskan tujuan dari dependensi di atas, dan mengapa saya memilih dependensi tersebut.

1. `go-chi` adalah library routing, membantu mengurai parameter permintaan dan merutekan permintaan. Alasan saya menggunakan library ini adalah karena tidak memiliki dependensi tidak langsung. Chi adalah library routing sederhana dan melakukan pekerjaannya dengan baik.

2. `pgx` adalah klien PostgreSQL. Saya tidak menggunakan `database/sql` standar karena biasanya saya menginginkan fitur PostgreSQL yang tidak tersedia pada driver database dengan denominator terendah.

3. `ulid` adalah library untuk menghasilkan dan membangun [ULID](https://github.com/ulid/spec). Saya membutuhkan ini karena saya ingin kunci utama domain diurutkan secara leksikografis tetapi juga stateless dan unik.

4. `null` adalah library yang biasanya saya gunakan untuk menghindari pengecualian pointer `nil` dan menghindari penggunaan pointer ketika saya dapat menggunakan nilai sebagai gantinya. Dependensi ini bersifat opsional.

5. `yaml` adalah library untuk mengurai file YAML. Ini juga merupakan dependensi opsional. Jika Anda tidak menyukai YAML, Anda dapat mengubahnya menjadi apa pun yang Anda inginkan.

6. `squirrel` adalah sebuah library untuk mengenerate kata-kata dalam bahasa SQL yang sesungguhnya😁.

## Cara Menavigasi Proyek Ini

### Paket `main`

Di root proyek ini ada paket `main` yang memiliki `main.go`. Ini adalah tempat di mana Anda meletakkan program utama Anda serta konfigurasi runtime.

### Modul-modul

Di dalam root proyek ada _modul_ yang diimplementasikan sebagai paket Go. Paket adalah batasan dalam Go. Semua yang ada di dalam modul ini diisolasi. Di dalam modul ini __*no rules*__ tentang bagaimana Anda mengatur file. Namun, dalam proyek saya, saya biasanya memiliki ini:

1. __Objek domain__. Ini mendefinisikan objek yang mempertahankan statusnya dan dipertahankan. Dalam contoh ini ada di file `todo_item.go`. Di dalam proyek ada `todo_item_json.go` ini hanya untuk serialisasi. Saya lebih suka memisahkan representasi JSON di file yang berbeda.

2. __Repositori__. Ini adalah abstraksi di mana Anda dapat mengambil dan menyimpan objek domain Anda. Saya hanya menamainya 'repository' karena _agak_ mirip dengan pola repositori tetapi saya mengimplementasikannya dengan fungsi murni di `repo.go`.

3. __Penyimpanan__. Ini hanyalah tempat di mana Anda terhubung, membaca, dan menulis di repositori Anda. Dalam contoh ini, sangat sederhana, hanya objek `pool` global yang mewakili kumpulan koneksi ke instance PostgreSQL. Itu ada di `db.go`.

4. __Model baca__. Ini adalah tipe dan struktur untuk mewakili sebagian atau agregasi data dari penyimpanan. Anda __*cannot modify*__ model baca karena seperti namanya, ini untuk membaca.

5. __Layanan__. Ini adalah file dengan fungsi yang mendefinisikan _batas transaksi_. Layanan ini agnostik dengan protokol. __Tidak boleh ada data terkait protokol__ di sini seperti Kode Respons HTTP. Juga diimplementasikan dalam fungsi murni di `service.go`.

6. __Protokol__. Ini adalah tempat di mana Anda meletakkan penangan *(handlers)* untuk permintaan Anda. Saya tidak meletakkannya di paket terpisah `handlers` karena saya tidak ingin menggunakan sesuatu seperti `CreateTodoItemHandler` dan sebagai gantinya saya hanya dapat menggunakannya langsung pada parameter function dari router tersebut seperti `route.Get("/", func() {...})`. Di dalam modul ada `protokol.go`. Setiap rute untuk modul ini ada di sini. Ini akan mengekspor objek `*chi.Mux` yang mengimplementasikan antarmuka `chi.Router`. Ini mengisolasi perubahan apa pun ke subrouter ke modul ini. Di `main.go` Anda hanya dapat memasang router.

> **Catatan**
> Sebelumnya, sebagian besar fungsi diekspor. Pada versi ini, saya telah membuat fungsi seprivat mungkin.

Fungsi yang diekspor di setiap area modul adalah sebagai berikut:

- `SetPool()` untuk menyiapkan kumpulan koneksi database global.
- `Router()` untuk mengekspor objek `chi.Mux` untuk dipasang.
- Objek domain, ini opsional. Jika Anda ingin menyembunyikan dan mengisolasi objek domain Anda, maka Anda dapat membuatnya privat.

## Pengujian/testing dan *Fake*

Saya jarang menggunakan mock. Baca alasannya [di sini](https://joeblu.com/blog/2023_06_mocks/). Saya menggunakan *fake*. Hal baik tentang Go adalah memungkinkan tag build yang akan memungkinkan kompilasi bersyarat file. Karena saya menggunakan file untuk menandai batas dan tanggung jawab, ini bekerja dengan baik.

Lihatlah `repo.go` dan `repo_fake.go` untuk perbandingan.

Untuk membangun atau menjalankan program dengan implementasi palsu, Anda dapat menggunakan parameter `-tags`, dan versi palsu akan dikompilasi sebagai gantinya.

```sh
# Build or run
go build -tags=fake 
```

## Konfigurasi

### *Rationale*

Versi awal program ini tidak memiliki file konfigurasi dan semua parameter di-hardcode. Pada awalnya saya pikir itu sudah cukup. Namun, saya pikir memberikan contoh tentang bagaimana mengimplementasikan server yang mematuhi aturan [12 Factor App](https://12factor.net/) adalah penting.

### Implementasi

Lihat `config.go`. File ini berisi kode untuk mengurai file konfigurasi dari dua sumber: variabel lingkungan dan file konfigurasi. File konfigurasi memiliki prioritas. Ini adalah variabel lingkungan, jalur kunci file konfigurasi, dan nilai default.

| Variabel Lingkungan  | Jalur kunci YAML  | Nilai default | Deskripsi          |
|-----------------------|---------------|---------------|----------------------|
| `KAD_LISTEN_HOST`     | `listen.host` | "127.0.0.1"   | Alamat Server       |
| `KAD_LISTEN_PORT`     | `listen.port` | 8080          | Port Server         |
| `KAD_DB_HOST`         | `db.host`     | "127.0.0.1"   | Host Postgres       |
| `KAD_DB_PORT`         | `db.port`     | 5432          | Port Postgres       |
| `KAD_DB_NAME`         | `db.dbname`  | "todo"        | Nama Database       |
| `KAD_DB_USERNAME`         | `db.username`  | "John"        | Username Database       |
| `KAD_DB_PASSWORD`         | `db.password`  | "example"        | Password Database       |
| `KAD_DB_SSL`          | `db.ssl_mode` | "disable"     | Mode SSL            |

Nilai default, jika kita mengungkapkannya dalam file konfigurasi adalah sebagai berikut.

```yaml
listen:
  host: 127.0.0.1
  port: 8080

db:
  dbname: todo
  host: 127.0.0.1
  port: 5432 
  ssl_mode: disable
  username: John
  password: example
```

### Lokasi file konfigurasi

Program akan mencari `config.yaml` di direktori kerja saat ini, atau Anda dapat memberikan flag `-c` untuk memaksa program menggunakan nama file konfigurasi Anda sendiri. Misalnya, Anda dapat menjalankannya dengan sesuatu seperti ini:

```sh
./app -c someconfig.yml
```

## Ringkasan

Proyek ini adalah heuristik, bukan panduan atau 'framework' struktur. Ini untuk menunjukkan bahwa Anda dapat memiliki struktur kode dan arsitektur yang masuk akal dengan tetap berpegang pada kesederhanaan Go.


## Cara build

```sh
go build -ldflags "-s -w" -o myApp

# atau pada windows
GOOS=windows GOARCH=amd64 go build -o app.exe -ldflags="-s -w"
```

## LICENSE

```
Copyright (c) 2025 Baroen Sudarman

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
```