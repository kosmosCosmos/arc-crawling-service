package doubanClient

import "fmt"

type DoubanServiceApiService service

func (d *DoubanServiceApiService) Hello(msg string) error {
	fmt.Println(d.client.cfg.ID)
	fmt.Println(d.client.cfg.Header)
	fmt.Println(msg)
	return nil
}
