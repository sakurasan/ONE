package models

import (
	//"os"
	// "path"
	"strconv"
	"strings"
	"time"

	// "github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	// _ "github.com/mattn/go-sqlite3"
	// "database/sql"
	_ "github.com/go-sql-driver/mysql"
)

const (
	// // 设置数据库路径
	// _DB_NAME = "data/beeblog.db"
	// 设置数据库名称
	_MYSQL_DRIVER = "mysql"

	cnn = "root:oneisall@/demo?charset=utf8&loc=Asia%2FShanghai"
)

// 分类
type Category struct {
	Id              int64
	Title           string
	Created         time.Time `orm:"index"`
	Views           int64     `orm:"index"`
	TopicTime       time.Time `orm:"index"`
	TopicCount      int64
	TopicLastUserId int64
}

// 文章
type Topic struct {
	Id              int64
	Uid             int64
	Title           string
	Category        string
	Labels          string
	Summary         string `orm:"size(1000)"`
	Content         string `orm:"size(5000)"`
	Tags            []*Tag `orm:"rel(m2m)"`
	Image           string
	Markdown        string
	Html            string
	Attachment      string
	Created         time.Time `orm:"index"`
	Updated         time.Time `orm:"index"`
	Views           int64     `orm:"index"`
	Author          string
	ReplyTime       time.Time `orm:"index"`
	ReplyCount      int64
	ReplyLastUserId int64
	Tflag           int
	Type            string
}

//标签
type Tag struct {
	Id    int64
	Name  string   `orm:"index"`
	Topic []*Topic `orm:"reverse(many)"`
}

//评论
type Comment struct {
	Id      int64
	Tid     int64
	Name    string
	Email   string
	Comment string    `orm:size(1000)`
	Created time.Time `orm:"index"`
}

func RegisterDB() {

	// 注册模型
	orm.RegisterModel(new(Category), new(Topic), new(Comment), new(Tag))
	// 注册驱动（“mysql”）
	orm.RegisterDriver("mysql", orm.DRMySQL)
	// 注册默认数据库
	orm.RegisterDataBase("default", _MYSQL_DRIVER, cnn, 10) //别名，名称，路径，最大连接数
}

func AddCategory(name string) error {
	o := orm.NewOrm()

	cate := &Category{
		Title:     name,
		Created:   time.Now(),
		TopicTime: time.Now(),
	}

	// 查询数据
	qs := o.QueryTable("category")
	err := qs.Filter("title", name).One(cate)
	if err == nil {
		return err
	}

	// 插入数据
	_, err = o.Insert(cate)
	if err != nil {
		return err
	}

	return nil
}

func DeleteCategory(id string) error {
	cid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	o := orm.NewOrm()

	cate := &Category{Id: cid}
	_, err = o.Delete(cate)
	return err
}

func GetAllCategories() ([]*Category, error) {
	o := orm.NewOrm()

	cates := make([]*Category, 0)

	qs := o.QueryTable("category")
	_, err := qs.All(&cates)
	return cates, err
}

/*** 添加 轻博，短博客 ***/
func AddTholeaf(summary, image string) error {

	o := orm.NewOrm()

	topic := &Topic{
		Summary:   summary,
		Image:     image,
		Created:   time.Now(),
		Updated:   time.Now(),
		ReplyTime: time.Now(),
		Tflag:     1,
	}
	_, err := o.Insert(topic)
	if err != nil {
		return err
	}
	return err
}

/***文章的 添加，获取所有文章，获取单篇文章内容，文章修改，文章删除***/
func AddTopic(title, category, label, summary, content, markdown, html string) error {
	o := orm.NewOrm()

	strs := strings.Split(label, " ")
	tags := []*Tag{}
	for _, v := range strs {
		tags = append(tags, Tag{Name: strings.TrimSpace(v)}.GetOrNewTag())
	}

	label = "$" + strings.Join(strings.Split(label, " "), "#$") + "#" // 切成切片，再用 $# 粘起来成字符

	topic := &Topic{
		Title:     title,
		Category:  category,
		Labels:    label,
		Tags:      tags,
		Summary:   summary,
		Content:   content,
		Markdown:  markdown,
		Html:      html,
		Created:   time.Now(),
		Updated:   time.Now(),
		ReplyTime: time.Now(),
	}

	_, err := o.Insert(topic)
	if err != nil {
		return err
	}

	m2m := o.QueryM2M(topic, "Tags") //更新文章标签对应关系
	m2m.Clear()
	m2m.Add(topic.Tags)

	//更新分类列表
	cate := new(Category)
	qs := o.QueryTable("category")
	err = qs.Filter("title", category).One(cate)
	if err == nil {
		cate.TopicCount++
		_, err = o.Update(cate)
	}

	return err
}

func GetAllTopics(cate string, isDesc, isTflag bool) ([]*Topic, error) {
	o := orm.NewOrm()

	topics := make([]*Topic, 0)

	qs := o.QueryTable("topic")

	var err error
	if isDesc {
		if len(cate) > 0 {
			qs = qs.Filter("Category", cate)
		}

		if isTflag {
			qs = qs.Filter("Tflag", 1)
			_, err = qs.OrderBy("-created").All(&topics)
		} else {
			qs = qs.Filter("Tflag", "")
			_, err = qs.OrderBy("-created").All(&topics)
		}

	} else {
		_, err = qs.All(&topics)
	}

	return topics, err

}

func GetAllTopiclist() ([]*Topic, error) {
	o := orm.NewOrm()

	topics := make([]*Topic, 0)

	qs := o.QueryTable("topic")

	qs = qs.Filter("Tflag", 0)
	_, err := qs.OrderBy("-created").All(&topics)

	return topics, err

}

func GetTopic(tid string) (*Topic, error) {
	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return nil, err
	}

	o := orm.NewOrm()

	topic := new(Topic)

	qs := o.QueryTable("topic")
	err = qs.Filter("id", tidNum).One(topic)
	if err != nil {
		return nil, err
	}

	topic.Views++
	_, err = o.Update(topic)

	topic.Labels = strings.Replace(
		strings.Replace(topic.Labels, "#", " ", -1), "$", "", -1)

	return topic, nil
}

func ModifyTopic(tid, title, category, summary, content, label string) error {
	o := orm.NewOrm()
	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return err
	}
	a := &Topic{Id: tidNum}
	old := Topic{Id: tidNum}
	_, _ = o.LoadRelated(&old, "Tags")

	strs := strings.Split(label, " ")                                 //以空格切割，切成slice
	label = "$" + strings.Join(strings.Split(label, " "), "#$") + "#" // 切成切片，再用 $# 粘起来成字符

	tags := []*Tag{}
	for _, v := range strs {
		tags = append(tags, Tag{Name: strings.TrimSpace(v)}.GetOrNewTag())
	}

	//分类
	var oldCate string

	topic := &Topic{Id: tidNum}

	err = o.Read(topic)
	if err == nil {
		//当查到，返回的 err为 nil
		oldCate = topic.Category //储存旧分类
		topic.Title = title
		topic.Category = category //新分类
		topic.Summary = summary
		topic.Content = content
		topic.Labels = label
		topic.Tags = tags
		topic.Updated = time.Now()
		_, err = o.Update(topic)
		if err != nil {
			return err
		}
	}

	//更新分类
	if len(oldCate) > 0 { //存在旧分类
		cate := new(Category)
		qs := o.QueryTable("category")
		err := qs.Filter("title", oldCate).One(cate) //查询文章分类是否在分类表中
		if err == nil {
			cate.TopicCount-- //查询存在,err为空 减一
			_, err = o.Update(cate)
		}
	}

	cate := new(Category)
	qs := o.QueryTable("category")
	err = qs.Filter("Title", category).One(cate) //新分类是否在 分类列表中
	if err == nil {
		cate.TopicCount++ //存在 加一
		_, err = o.Update(cate)
	}

	a.Tags = tags
	m2m := o.QueryM2M(topic, "Tags")

	//标签插入
VI:
	for _, vi := range a.Tags {
		for _, vj := range old.Tags {
			if vi.Id == vj.Id {
				continue VI
			}
		}
		m2m.Add(vi)
	}

	//标签删除的
VD:
	for _, vi := range old.Tags {
		for _, vj := range a.Tags {
			if vi.Id == vj.Id {
				continue VD
			}
		}
		m2m.Remove(vi)
	}

	return nil
}

func DeleteTopic(tid string) error {
	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return err
	}
	var oldCate string
	o := orm.NewOrm()

	topic := &Topic{Id: tidNum}
	if o.Read(topic) == nil {
		oldCate = topic.Category
		_, err = o.Delete(topic)
		if err != nil {
			return err
		}
	}

	if len(oldCate) > 0 {
		cate := new(Category)
		qs := o.QueryTable("category")
		err = qs.Filter("title", oldCate).One(cate)
		if err == nil {
			cate.TopicCount--
			_, err = o.Update(cate)
		}
	}

	_, err = o.Delete(topic)

	return err
}

/******************* 评 论 ****************/
func AddReply(tid, nickname, email, comment string) error {
	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return err
	}
	o := orm.NewOrm()

	reply := &Comment{
		Tid:     tidNum,
		Name:    nickname,
		Email:   email,
		Comment: comment,
		Created: time.Now(),
	}

	_, err = o.Insert(reply)
	if err != nil {
		return err
	}

	//更新 评论总数
	listes := make([]*Comment, 0)

	qs := o.QueryTable("comment")

	_, err = qs.Filter("tid", tidNum).All(&listes)
	if err != nil {
		return err
	}

	topic := new(Topic)
	qs = o.QueryTable("topic")
	err = qs.Filter("id", tidNum).One(topic)
	if err == nil {
		topic.ReplyCount = int64(len(listes))
		_, err = o.Update(topic)
	}

	return err

	// var lists []orm.ParamsList

	// replycnt, err := o.QueryTable("comment").Filter("tid", tidNum).ValuesList(&lists)
	// if err != nil {
	// 	return nil
	// }
	// topic := &Topic{Id: tidNum}
	// err = o.Read("topic")
	// if err == nil {
	// 	topic.ReplyCount = replycnt
	// 	o.Update(topic)
	// }

	// return err

}

func GetAllReplies(tid string) (replies []*Comment, err error) {
	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return nil, err
	}
	o := orm.NewOrm()

	replies = make([]*Comment, 0)

	qs := o.QueryTable("comment")
	_, err = qs.Filter("Tid", tidNum).All(&replies)

	return replies, err

}

func DeleteReply(rid string) error {
	ridNum, err := strconv.ParseInt(rid, 10, 64)
	if err != nil {
		return err
	}

	o := orm.NewOrm()
	var tidNum int64

	reply := &Comment{Id: ridNum}
	if o.Read(reply) == nil {
		tidNum = reply.Tid //把评论对应的文章ID取出来

		_, err = o.Delete(reply)
		if err != nil {
			return err
		}
	}

	//更新评论总数
	// var lists []orm.ParamsList  //法二行不通

	// replycnt, err := o.QueryTable("comment").Filter("tid", tidNum).ValuesList(&lists)
	// if err != nil {
	// 	return nil
	// }
	listes := make([]*Comment, 0)

	qs := o.QueryTable("comment")

	_, err = qs.Filter("tid", tidNum).All(&listes)
	if err != nil {
		return err
	}

	topic := &Topic{Id: tidNum}
	if o.Read(topic) == nil {
		topic.ReplyCount = int64(len(listes))
		o.Update(topic)
	}

	return err

}

func GetAllTags() ([]*Tag, error) {
	tags := make([]*Tag, 0)
	o := orm.NewOrm()
	qs := o.QueryTable("tag").Distinct()
	_, err := qs.All(&tags)
	return tags, err
}

func Tag2topic(tag string) (topics []*Topic, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable("topic")
	topics = make([]*Topic, 0)
	if len(tag) > 0 {
		qs = qs.Filter("labels__contains", "$"+tag+"#") //用到beego orm高级查询 包含
	}
	_, err = qs.OrderBy("-created").All(&topics)
	return topics, err

}

func (t Tag) GetOrNewTag() *Tag {
	o := orm.NewOrm()
	_, _, _ = o.ReadOrCreate(&t, "Name")
	return &t
}
