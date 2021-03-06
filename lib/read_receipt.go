package readraptor

import (
	"time"

	"github.com/coopernurse/gorp"
)

type ReadReceipt struct {
	Id            int64     `db:"id"`
	Created       time.Time `db:"created_at"`
	ArticleId int64     `db:"article_id"`
	ReaderId      int64     `db:"reader_id"`
}

func TrackReadReceipt(dbmap *gorp.DbMap, account *Account, key, reader string) error {
	cid, err := InsertArticle(dbmap, account.Id, key)
	if err != nil {
		return err
	}

	vid, err := InsertReader(dbmap, account.Id, reader)
	if err != nil {
		return err
	}

	_, err = InsertReadReceipt(dbmap, cid, vid)
	if err != nil {
		return err
	}

	return err
}

func InsertReadReceipt(dbmap *gorp.DbMap, articleId, readerId int64) (int64, error) {
	id, err := dbmap.SelectNullInt(`
        with s as (
            select id from read_receipts where article_id = $1 and reader_id = $2
        ), i as (
            insert into read_receipts ("article_id", "reader_id", "created_at")
            select $1, $2, $3
            where not exists (select 1 from s)
            returning id
        )
        select id from i union all select id from s;
    `, articleId,
		readerId,
		time.Now(),
	)
	if err != nil {
		return -1, err
	}

	iid, err := id.Value()

	return iid.(int64), err
}

func UnreadArticles(dbmap *gorp.DbMap, readerId int64) (keys []Article, err error) {
	_, err = dbmap.Select(&keys, `
        select articles.* from
            (select article_id from expected_readers where reader_id = $1
                except all
             select article_id from read_receipts where reader_id = $1) unread_articles
        inner join articles on articles.id = unread_articles.article_id;`, readerId)

	return
}
