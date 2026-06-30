import fs from 'node:fs/promises'
import path from 'node:path'

const LOCALES_DIR = path.resolve('src/i18n/locales')

function stableStringify(obj) {
  return JSON.stringify(obj, null, 2) + '\n'
}

const newKeys = {
  en: {
    'Alibaba OSS': 'Alibaba OSS',
    'Allow dashboard users to upload temporary session attachments in Playground.':
      'Allow dashboard users to upload temporary session attachments in Playground.',
    'Attachment TTL (hours)': 'Attachment TTL (hours)',
    'Attachments expire after this many hours.':
      'Attachments expire after this many hours.',
    'Choose where temporary Playground attachment files are stored.':
      'Choose where temporary Playground attachment files are stored.',
    'Cleanup batch size': 'Cleanup batch size',
    'Cleanup interval (minutes)': 'Cleanup interval (minutes)',
    'Each line becomes one allowed MIME type for Playground attachment uploads.':
      'Each line becomes one allowed MIME type for Playground attachment uploads.',
    'Enable Playground attachments': 'Enable Playground attachments',
    'Expired attachments are scanned on this schedule.':
      'Expired attachments are scanned on this schedule.',
    'Filesystem directory used when the local storage driver is selected.':
      'Filesystem directory used when the local storage driver is selected.',
    'Leave blank to keep the existing secret':
      'Leave blank to keep the existing secret',
    'Local filesystem': 'Local filesystem',
    'Local storage path': 'Local storage path',
    'Max file size (bytes)': 'Max file size (bytes)',
    'Max files per message': 'Max files per message',
    'Max files per session': 'Max files per session',
    'Maximum allowed size for a single attachment upload.':
      'Maximum allowed size for a single attachment upload.',
    'Maximum number of active attachments kept for one Playground session.':
      'Maximum number of active attachments kept for one Playground session.',
    'Maximum number of attachments allowed per send.':
      'Maximum number of attachments allowed per send.',
    'Number of expired attachments processed per cleanup run.':
      'Number of expired attachments processed per cleanup run.',
    'One MIME type per line': 'One MIME type per line',
    'OSS access key ID': 'OSS access key ID',
    'OSS access key secret': 'OSS access key secret',
    'OSS bucket': 'OSS bucket',
    'OSS endpoint': 'OSS endpoint',
    'OSS object prefix': 'OSS object prefix',
    'OSS region': 'OSS region',
    'Playground Attachments': 'Playground Attachments',
    'Reference TTL (seconds)': 'Reference TTL (seconds)',
    'Save Playground attachment settings':
      'Save Playground attachment settings',
    'Sensitive values are hidden when loaded. Enter a new value only when you want to replace the current credential.':
      'Sensitive values are hidden when loaded. Enter a new value only when you want to replace the current credential.',
    'Selected model may not support this attachment type':
      'Selected model may not support this attachment type',
    'Selected model only supports image attachments':
      'Selected model only supports image attachments',
    'Select storage driver': 'Select storage driver',
    'Signed read links remain valid for this many seconds.':
      'Signed read links remain valid for this many seconds.',
    'Storage driver': 'Storage driver',
    'Allowed MIME types': 'Allowed MIME types',
  },
  zh: {
    'Alibaba OSS': '阿里云 OSS',
    'Allow dashboard users to upload temporary session attachments in Playground.':
      '允许已登录的控制台用户在 Playground 中上传临时会话附件。',
    'Attachment TTL (hours)': '附件存活时长（小时）',
    'Attachments expire after this many hours.':
      '附件会在这么多小时后过期。',
    'Choose where temporary Playground attachment files are stored.':
      '选择临时 Playground 附件文件的存储位置。',
    'Cleanup batch size': '清理批次大小',
    'Cleanup interval (minutes)': '清理间隔（分钟）',
    'Each line becomes one allowed MIME type for Playground attachment uploads.':
      '每一行都会作为一个允许的 Playground 附件 MIME 类型。',
    'Enable Playground attachments': '启用 Playground 附件',
    'Expired attachments are scanned on this schedule.':
      '系统会按此计划扫描过期附件。',
    'Filesystem directory used when the local storage driver is selected.':
      '选择本地存储驱动时使用的文件系统目录。',
    'Leave blank to keep the existing secret': '留空以保留现有密钥',
    'Local filesystem': '本地文件系统',
    'Local storage path': '本地存储路径',
    'Max file size (bytes)': '最大文件大小（字节）',
    'Max files per message': '每条消息最大文件数',
    'Max files per session': '每个会话最大文件数',
    'Maximum allowed size for a single attachment upload.':
      '单个附件上传允许的最大大小。',
    'Maximum number of active attachments kept for one Playground session.':
      '单个 Playground 会话允许保留的活动附件数量上限。',
    'Maximum number of attachments allowed per send.':
      '每次发送允许的附件数量上限。',
    'Number of expired attachments processed per cleanup run.':
      '每次清理运行处理的过期附件数量。',
    'One MIME type per line': '每行一个 MIME 类型',
    'OSS access key ID': 'OSS Access Key ID',
    'OSS access key secret': 'OSS Access Key Secret',
    'OSS bucket': 'OSS Bucket',
    'OSS endpoint': 'OSS Endpoint',
    'OSS object prefix': 'OSS 对象前缀',
    'OSS region': 'OSS 区域',
    'Playground Attachments': 'Playground 附件',
    'Reference TTL (seconds)': '引用有效期（秒）',
    'Save Playground attachment settings': '保存 Playground 附件设置',
    'Sensitive values are hidden when loaded. Enter a new value only when you want to replace the current credential.':
      '敏感字段在加载时会被隐藏。只有在需要替换当前凭证时才输入新值。',
    'Selected model may not support this attachment type':
      '当前所选模型可能不支持此附件类型',
    'Selected model only supports image attachments':
      '当前所选模型仅支持图片附件',
    'Select storage driver': '选择存储驱动',
    'Signed read links remain valid for this many seconds.':
      '签名读取链接会在这么多秒内保持有效。',
    'Storage driver': '存储驱动',
    'Allowed MIME types': '允许的 MIME 类型',
  },
  fr: {
    'Alibaba OSS': 'Alibaba OSS',
    'Allow dashboard users to upload temporary session attachments in Playground.':
      'Permettre aux utilisateurs connectés du tableau de bord de téléverser des pièces jointes de session temporaires dans Playground.',
    'Attachment TTL (hours)': 'TTL des pièces jointes (heures)',
    'Attachments expire after this many hours.':
      'Les pièces jointes expirent après ce nombre d’heures.',
    'Choose where temporary Playground attachment files are stored.':
      'Choisissez où stocker les fichiers temporaires des pièces jointes Playground.',
    'Cleanup batch size': 'Taille du lot de nettoyage',
    'Cleanup interval (minutes)': 'Intervalle de nettoyage (minutes)',
    'Each line becomes one allowed MIME type for Playground attachment uploads.':
      'Chaque ligne devient un type MIME autorisé pour les téléversements de pièces jointes Playground.',
    'Enable Playground attachments': 'Activer les pièces jointes Playground',
    'Expired attachments are scanned on this schedule.':
      'Les pièces jointes expirées sont analysées selon cette fréquence.',
    'Filesystem directory used when the local storage driver is selected.':
      'Répertoire du système de fichiers utilisé lorsque le pilote de stockage local est sélectionné.',
    'Leave blank to keep the existing secret':
      'Laissez vide pour conserver le secret existant',
    'Local filesystem': 'Système de fichiers local',
    'Local storage path': 'Chemin de stockage local',
    'Max file size (bytes)': 'Taille maximale du fichier (octets)',
    'Max files per message': 'Nombre max. de fichiers par message',
    'Max files per session': 'Nombre max. de fichiers par session',
    'Maximum allowed size for a single attachment upload.':
      'Taille maximale autorisée pour un seul téléversement de pièce jointe.',
    'Maximum number of active attachments kept for one Playground session.':
      'Nombre maximal de pièces jointes actives conservées pour une session Playground.',
    'Maximum number of attachments allowed per send.':
      'Nombre maximal de pièces jointes autorisées par envoi.',
    'Number of expired attachments processed per cleanup run.':
      'Nombre de pièces jointes expirées traitées à chaque nettoyage.',
    'One MIME type per line': 'Un type MIME par ligne',
    'OSS access key ID': 'ID de clé d’accès OSS',
    'OSS access key secret': 'Secret de clé d’accès OSS',
    'OSS bucket': 'Bucket OSS',
    'OSS endpoint': 'Point de terminaison OSS',
    'OSS object prefix': 'Préfixe d’objet OSS',
    'OSS region': 'Région OSS',
    'Playground Attachments': 'Pièces jointes Playground',
    'Reference TTL (seconds)': 'TTL de référence (secondes)',
    'Save Playground attachment settings':
      'Enregistrer les paramètres des pièces jointes Playground',
    'Sensitive values are hidden when loaded. Enter a new value only when you want to replace the current credential.':
      'Les valeurs sensibles sont masquées au chargement. Saisissez une nouvelle valeur uniquement si vous souhaitez remplacer l’identifiant actuel.',
    'Selected model may not support this attachment type':
      "Le modèle sélectionné ne prend peut-être pas en charge ce type de pièce jointe",
    'Selected model only supports image attachments':
      "Le modèle sélectionné prend uniquement en charge les pièces jointes image",
    'Select storage driver': 'Sélectionner un pilote de stockage',
    'Signed read links remain valid for this many seconds.':
      'Les liens de lecture signés restent valides pendant ce nombre de secondes.',
    'Storage driver': 'Pilote de stockage',
    'Allowed MIME types': 'Types MIME autorisés',
  },
  ja: {
    'Alibaba OSS': 'Alibaba OSS',
    'Allow dashboard users to upload temporary session attachments in Playground.':
      'ダッシュボードにログインしたユーザーが Playground で一時的なセッション添付ファイルをアップロードできるようにします。',
    'Attachment TTL (hours)': '添付ファイル TTL（時間）',
    'Attachments expire after this many hours.':
      '添付ファイルはこの時間数が経過すると期限切れになります。',
    'Choose where temporary Playground attachment files are stored.':
      '一時的な Playground 添付ファイルの保存先を選択します。',
    'Cleanup batch size': 'クリーンアップのバッチサイズ',
    'Cleanup interval (minutes)': 'クリーンアップ間隔（分）',
    'Each line becomes one allowed MIME type for Playground attachment uploads.':
      '各行が Playground 添付アップロードで許可される MIME タイプになります。',
    'Enable Playground attachments': 'Playground 添付を有効化',
    'Expired attachments are scanned on this schedule.':
      '期限切れの添付ファイルはこの間隔でスキャンされます。',
    'Filesystem directory used when the local storage driver is selected.':
      'ローカルストレージドライバー選択時に使用するファイルシステムディレクトリです。',
    'Leave blank to keep the existing secret': '既存のシークレットを保持するには空欄のままにします',
    'Local filesystem': 'ローカルファイルシステム',
    'Local storage path': 'ローカル保存パス',
    'Max file size (bytes)': '最大ファイルサイズ（バイト）',
    'Max files per message': '1 メッセージあたりの最大ファイル数',
    'Max files per session': '1 セッションあたりの最大ファイル数',
    'Maximum allowed size for a single attachment upload.':
      '1 つの添付ファイルアップロードで許可される最大サイズです。',
    'Maximum number of active attachments kept for one Playground session.':
      '1 つの Playground セッションで保持されるアクティブ添付の最大数です。',
    'Maximum number of attachments allowed per send.':
      '1 回の送信で許可される添付ファイルの最大数です。',
    'Number of expired attachments processed per cleanup run.':
      '各クリーンアップ実行で処理される期限切れ添付ファイル数です。',
    'One MIME type per line': '1 行につき 1 つの MIME タイプ',
    'OSS access key ID': 'OSS アクセスキー ID',
    'OSS access key secret': 'OSS アクセスキーシークレット',
    'OSS bucket': 'OSS バケット',
    'OSS endpoint': 'OSS エンドポイント',
    'OSS object prefix': 'OSS オブジェクトプレフィックス',
    'OSS region': 'OSS リージョン',
    'Playground Attachments': 'Playground 添付',
    'Reference TTL (seconds)': '参照 TTL（秒）',
    'Save Playground attachment settings': 'Playground 添付設定を保存',
    'Sensitive values are hidden when loaded. Enter a new value only when you want to replace the current credential.':
      '機密値は読み込み時に非表示になります。現在の認証情報を置き換える場合のみ新しい値を入力してください。',
    'Selected model may not support this attachment type':
      '選択したモデルはこの添付ファイル形式をサポートしていない可能性があります',
    'Selected model only supports image attachments':
      '選択したモデルは画像添付のみをサポートしています',
    'Select storage driver': 'ストレージドライバーを選択',
    'Signed read links remain valid for this many seconds.':
      '署名付き読み取りリンクはこの秒数の間有効です。',
    'Storage driver': 'ストレージドライバー',
    'Allowed MIME types': '許可する MIME タイプ',
  },
  ru: {
    'Alibaba OSS': 'Alibaba OSS',
    'Allow dashboard users to upload temporary session attachments in Playground.':
      'Разрешить пользователям панели управления загружать временные вложения сессии в Playground.',
    'Attachment TTL (hours)': 'TTL вложения (часы)',
    'Attachments expire after this many hours.':
      'Вложения истекают через указанное количество часов.',
    'Choose where temporary Playground attachment files are stored.':
      'Выберите, где хранить временные файлы вложений Playground.',
    'Cleanup batch size': 'Размер пакета очистки',
    'Cleanup interval (minutes)': 'Интервал очистки (минуты)',
    'Each line becomes one allowed MIME type for Playground attachment uploads.':
      'Каждая строка становится одним разрешённым MIME-типом для загрузки вложений Playground.',
    'Enable Playground attachments': 'Включить вложения Playground',
    'Expired attachments are scanned on this schedule.':
      'Просроченные вложения проверяются по этому расписанию.',
    'Filesystem directory used when the local storage driver is selected.':
      'Каталог файловой системы, используемый при выборе локального драйвера хранения.',
    'Leave blank to keep the existing secret':
      'Оставьте пустым, чтобы сохранить существующий секрет',
    'Local filesystem': 'Локальная файловая система',
    'Local storage path': 'Путь локального хранения',
    'Max file size (bytes)': 'Макс. размер файла (байты)',
    'Max files per message': 'Макс. файлов на сообщение',
    'Max files per session': 'Макс. файлов на сессию',
    'Maximum allowed size for a single attachment upload.':
      'Максимально допустимый размер одной загрузки вложения.',
    'Maximum number of active attachments kept for one Playground session.':
      'Максимальное количество активных вложений, сохраняемых для одной сессии Playground.',
    'Maximum number of attachments allowed per send.':
      'Максимальное количество вложений, разрешённых за одну отправку.',
    'Number of expired attachments processed per cleanup run.':
      'Количество просроченных вложений, обрабатываемых за один запуск очистки.',
    'One MIME type per line': 'Один MIME-тип на строку',
    'OSS access key ID': 'ID ключа доступа OSS',
    'OSS access key secret': 'Секрет ключа доступа OSS',
    'OSS bucket': 'Бакет OSS',
    'OSS endpoint': 'Эндпоинт OSS',
    'OSS object prefix': 'Префикс объекта OSS',
    'OSS region': 'Регион OSS',
    'Playground Attachments': 'Вложения Playground',
    'Reference TTL (seconds)': 'TTL ссылки (секунды)',
    'Save Playground attachment settings':
      'Сохранить настройки вложений Playground',
    'Sensitive values are hidden when loaded. Enter a new value only when you want to replace the current credential.':
      'Чувствительные значения скрываются при загрузке. Введите новое значение только если хотите заменить текущие учётные данные.',
    'Selected model may not support this attachment type':
      'Выбранная модель может не поддерживать этот тип вложения',
    'Selected model only supports image attachments':
      'Выбранная модель поддерживает только вложения-изображения',
    'Select storage driver': 'Выберите драйвер хранения',
    'Signed read links remain valid for this many seconds.':
      'Подписанные ссылки для чтения остаются действительными столько секунд.',
    'Storage driver': 'Драйвер хранения',
    'Allowed MIME types': 'Разрешённые MIME-типы',
  },
  vi: {
    'Alibaba OSS': 'Alibaba OSS',
    'Allow dashboard users to upload temporary session attachments in Playground.':
      'Cho phép người dùng bảng điều khiển đã đăng nhập tải lên tệp đính kèm phiên tạm thời trong Playground.',
    'Attachment TTL (hours)': 'TTL tệp đính kèm (giờ)',
    'Attachments expire after this many hours.':
      'Tệp đính kèm sẽ hết hạn sau từng này giờ.',
    'Choose where temporary Playground attachment files are stored.':
      'Chọn nơi lưu các tệp đính kèm Playground tạm thời.',
    'Cleanup batch size': 'Kích thước lô dọn dẹp',
    'Cleanup interval (minutes)': 'Khoảng thời gian dọn dẹp (phút)',
    'Each line becomes one allowed MIME type for Playground attachment uploads.':
      'Mỗi dòng sẽ trở thành một kiểu MIME được phép cho tải lên tệp đính kèm Playground.',
    'Enable Playground attachments': 'Bật tệp đính kèm Playground',
    'Expired attachments are scanned on this schedule.':
      'Các tệp đính kèm hết hạn sẽ được quét theo lịch này.',
    'Filesystem directory used when the local storage driver is selected.':
      'Thư mục hệ thống tệp được dùng khi chọn trình điều khiển lưu trữ cục bộ.',
    'Leave blank to keep the existing secret':
      'Để trống để giữ lại bí mật hiện có',
    'Local filesystem': 'Hệ thống tệp cục bộ',
    'Local storage path': 'Đường dẫn lưu trữ cục bộ',
    'Max file size (bytes)': 'Kích thước tệp tối đa (byte)',
    'Max files per message': 'Số tệp tối đa mỗi tin nhắn',
    'Max files per session': 'Số tệp tối đa mỗi phiên',
    'Maximum allowed size for a single attachment upload.':
      'Kích thước tối đa cho phép của một lần tải lên tệp đính kèm.',
    'Maximum number of active attachments kept for one Playground session.':
      'Số lượng tệp đính kèm đang hoạt động tối đa được giữ cho một phiên Playground.',
    'Maximum number of attachments allowed per send.':
      'Số lượng tệp đính kèm tối đa được phép cho mỗi lần gửi.',
    'Number of expired attachments processed per cleanup run.':
      'Số tệp đính kèm hết hạn được xử lý trong mỗi lần dọn dẹp.',
    'One MIME type per line': 'Một MIME type trên mỗi dòng',
    'OSS access key ID': 'ID khoa truy cap OSS',
    'OSS access key secret': 'Bi mat khoa truy cap OSS',
    'OSS bucket': 'Bucket OSS',
    'OSS endpoint': 'Diem cuoi OSS',
    'OSS object prefix': 'Tiền tố đối tượng OSS',
    'OSS region': 'Vùng OSS',
    'Playground Attachments': 'Tệp đính kèm Playground',
    'Reference TTL (seconds)': 'TTL tham chiếu (giây)',
    'Save Playground attachment settings':
      'Lưu cài đặt tệp đính kèm Playground',
    'Sensitive values are hidden when loaded. Enter a new value only when you want to replace the current credential.':
      'Các giá trị nhạy cảm sẽ bị ẩn khi tải. Chỉ nhập giá trị mới khi bạn muốn thay thế thông tin xác thực hiện tại.',
    'Selected model may not support this attachment type':
      'Mô hình đã chọn có thể không hỗ trợ loại tệp đính kèm này',
    'Selected model only supports image attachments':
      'Mô hình đã chọn chỉ hỗ trợ tệp đính kèm hình ảnh',
    'Select storage driver': 'Chọn trình điều khiển lưu trữ',
    'Signed read links remain valid for this many seconds.':
      'Liên kết đọc đã ký sẽ còn hiệu lực trong từng này giây.',
    'Storage driver': 'Trình điều khiển lưu trữ',
    'Allowed MIME types': 'Các MIME type được phép',
  },
}

async function main() {
  let totalAdded = 0

  for (const [locale, trans] of Object.entries(newKeys)) {
    const filePath = path.join(LOCALES_DIR, `${locale}.json`)
    const json = JSON.parse(await fs.readFile(filePath, 'utf8'))

    let count = 0
    for (const [key, value] of Object.entries(trans)) {
      if (!Object.prototype.hasOwnProperty.call(json.translation, key)) {
        json.translation[key] = value
        count++
      } else if (json.translation[key] !== value) {
        json.translation[key] = value
        count++
      }
    }

    if (count > 0) {
      json.translation = Object.fromEntries(
        Object.entries(json.translation).sort(([a], [b]) => a.localeCompare(b))
      )
      await fs.writeFile(filePath, stableStringify(json), 'utf8')
    }

    console.log(`${locale}: ${count} translations applied`)
    totalAdded += count
  }

  console.log(`\nTotal: ${totalAdded} translations applied`)
}

main().catch((err) => {
  console.error(err)
  process.exitCode = 1
})
