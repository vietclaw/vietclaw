package agentfs_test

import (
	"strings"
	"testing"

	"vietclaw/internal/agentfs"
)

func sampleDetailedPersona() string {
	sections := []string{
		"## Vai trò & mục tiêu\nNhiệm vụ chính của agent này là nghiên cứu và so sánh dịch vụ VPS tại Việt Nam, tổng hợp bảng giá và đưa ra khuyến nghị rõ ràng cho người dùng.",
		"## Nhiệm vụ chính\nThu thập giá VPS từ nhà cung cấp uy tín, so sánh CPU/RAM/storage/bandwidth, ghi nguồn và ngày tra cứu.",
		"## Quy trình từng bước\n1) web_search với query cụ thể. 2) Chọn URL đáng tin. 3) web_fetch từng URL. 4) Tổng hợp bảng markdown.",
		"## Định dạng output\nLuôn trả bảng markdown + bullet tóm tắt + nguồn URL. Không bịa giá.",
		"## Quy tắc tool\nweb_search trước web_fetch. Không đoán domain.",
		"## Giới hạn\nKhông cam kết giá chính thức; báo user kiểm tra lại trên site nhà cung cấp.",
		"## Ví dụ task\nSo sánh gói VPS 2GB RAM của Viettel IDC, VNPT và FPT cho website nhỏ.",
	}
	return strings.Join(sections, "\n\n")
}

func sampleDetailedRequest() agentfs.CreateRequest {
	return agentfs.CreateRequest{
		ID:       "vps-analyst",
		Name:     "VPS Analyst",
		Language: "vi",
		Persona:  sampleDetailedPersona(),
		Tools:    []string{"web_search", "web_fetch"},
		Skills: []agentfs.SkillInput{{
			Name:         "market-research",
			Triggers:     []string{"VPS", "pricing"},
			Instructions: strings.Repeat("Thu thập giá từ nhiều nguồn, đối chiếu spec, ghi chú hạn chế dữ liệu và đề xuất 2-3 gói phù hợp theo ngân sách user. ", 3),
		}},
		ToolGuides: []agentfs.ToolGuideInput{
			{
				Tool:         "web_search",
				Triggers:     []string{"pricing", "compare"},
				Instructions: strings.Repeat("Dùng query tiếng Việt và tiếng Anh, ưu tiên từ khóa nhà cung cấp + năm. ", 4),
			},
			{
				Tool:         "web_fetch",
				Triggers:     []string{"URL"},
				Instructions: strings.Repeat("Chỉ fetch URL từ kết quả search hoặc user cung cấp; tối đa 5 URL mỗi lần. ", 4),
			},
		},
	}
}

func TestValidateCreateRequestRejectsShortPersona(t *testing.T) {
	req := sampleDetailedRequest()
	req.Persona = "Agent ngắn."
	err := agentfs.ValidateCreateRequest(req)
	if err == nil {
		t.Fatal("expected short persona error")
	}
}

func TestValidateCreateRequestRejectsMissingSections(t *testing.T) {
	req := sampleDetailedRequest()
	req.Persona = strings.Repeat("Một đoạn dài nhưng không có heading markdown đủ số lượng. ", 30)
	err := agentfs.ValidateCreateRequest(req)
	if err == nil {
		t.Fatal("expected missing sections error")
	}
}

func TestValidateCreateRequestAcceptsDetailedAgent(t *testing.T) {
	if err := agentfs.ValidateCreateRequest(sampleDetailedRequest()); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
}
