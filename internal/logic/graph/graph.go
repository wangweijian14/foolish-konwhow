package graph

import (
	"context"
	"fmt"
	"naivete/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type sGraph struct {
	Driver neo4j.DriverWithContext
	Dict   []string
	Cache  *gcache.Cache
}

func init() {
	service.RegisterGraph(New())
}
func New() *sGraph {
	// "bolt://127.0.0.1:7687", "neo4j", "admaster54322"
	uri := "bolt://127.0.0.1:7687"
	username := "neo4j"
	password := "admaster54322"
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		panic(err)
	}
	return &sGraph{
		Driver: driver,
		Dict:   []string{"tools", "frame", "technical_ability", "space", "method"},
		Cache:  gcache.New(),
	}
}

// 构建知识库
func (s *sGraph) BuildGraph(ctx context.Context) error {

	// 入库dict
	tools := []string{}
	// 字典
	dicts := []string{}
	for _, dict := range s.Dict {

		err := gfile.ReadLines(fmt.Sprintf("dict/%s.data", dict), func(line string) error {
			var err error
			if dict == "tools" {
				j, err := gjson.DecodeToJson(line)
				if err != nil {
					return err
				}
				// cache 实体关系边
				name := j.Get("name")
				method := j.Get("method")
				technical_ability := j.Get("technical_ability")
				frame := j.Get("frame")
				space := j.Get("space")
				s.cacheData(ctx, name.String()+"#belong_method", method)
				s.cacheData(ctx, name.String()+"#need_technical_ability", technical_ability)
				s.cacheData(ctx, name.String()+"#belongs_frame", frame)
				err = s.cacheData(ctx, name.String()+"#belongs_space", space)
				if err != nil {
					return err
				}
				tools = append(tools, name.String())
				_, err = s.Create(ctx, dict, j.Map())
				s.Cache.Set(ctx, name.String(), 1, 0)
			} else {
				_, err = s.Create(ctx, dict, map[string]any{
					"name": line,
				})
				dicts = append(dicts, line)
			}
			s.Cache.Set(ctx, line, 1, 0)
			return err
		})
		if err != nil {
			return err
		}
	}
	g.Log().Info(ctx, "build dict node done!")

	// cache
	dicts = append(dicts, tools...)
	s.Cache.Set(ctx, "tools", tools, 0)
	s.Cache.Set(ctx, "dicts", tools, 0)

	// '''创建实体关系边'''
	g.Log().Info(ctx, "[创建实体关系边-tools: ]", tools)
	for _, tool := range tools {
		belong_method, _ := s.Cache.Get(ctx, fmt.Sprintf("%s#belong_method", tool))
		need_technical_ability, _ := s.Cache.Get(ctx, fmt.Sprintf("%s#need_technical_ability", tool))
		belongs_frame, _ := s.Cache.Get(ctx, fmt.Sprintf("%s#belongs_frame", tool))
		belongs_space, _ := s.Cache.Get(ctx, fmt.Sprintf("%s#belongs_space", tool))
		g.Log().Info(ctx, "method:", belong_method.Strings())
		s.Create_relationship(ctx, "tools", "method", tool, belong_method.Strings(), "belong_method", "测试方法类型")
		s.Create_relationship(ctx, "tools", "need_technical_ability", tool, need_technical_ability.Strings(), "need_technical_ability", "需要掌握的技能")
		s.Create_relationship(ctx, "tools", "frame", tool, belongs_frame.Strings(), "belongs_frame", "属于框架类型")
		s.Create_relationship(ctx, "tools", "space", tool, belongs_space.Strings(), "belongs_space", "所属应用领域")

	}
	g.Log().Info(ctx, "build tool attr relationship done!")
	return nil
}

func (s *sGraph) Create(ctx context.Context, nodeName string, params map[string]any) (any, error) {
	session := s.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)
	p := ""
	for k, v := range params {
		p += fmt.Sprintf("%s:\"%v\"", k, v) + ","
	}

	cql := fmt.Sprintf("create (%v:%v{%v})", nodeName, nodeName, p[:len(p)-1])
	g.Log().Info(ctx, cql)
	res, err := session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx, cql, nil)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s *sGraph) ExecuteCql(ctx context.Context, cql string) (any, error) {
	session := s.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)
	g.Log().Info(ctx, cql)
	res, err := session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx, cql, nil)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s *sGraph) Create_relationship(ctx context.Context, start_node string, end_node, pname string, edges []string, rel_type, rel_name string) error {
	count := 0
	// # 去重处理
	for _, edge := range edges {
		// gvarP, _ := s.Cache.Get(ctx, pname)
		// gvarQ, _ := s.Cache.Get(ctx, edge)
		// if gvarP.Int() == 1 && gvarQ.Int() == 1 {
		query := fmt.Sprintf("match(p:%s),(q:%s) where p.name='%s'and q.name='%s' create (p)-[rel:%s{name:'%s'}]->(q)",
			start_node, end_node, pname, edge, rel_type, rel_name)
		g.Log().Info(ctx, "[insert] ", query)
		_, err := s.ExecuteCql(ctx, query)
		if err != nil {
			return err
		}
		count++
	}
	// }
	g.Log().Info(ctx, "[count]", count)

	return nil
}

func (s *sGraph) cacheData(ctx context.Context, k string, v *g.Var) error {
	if !v.IsEmpty() {
		err := s.Cache.Set(ctx, k, v.Strings(), 0)
		if err != nil {
			return err
		}
	}
	return nil
}
