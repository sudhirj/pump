package pump

import "testing"

func TestObject_IsCompletedBy(t *testing.T) {
	type fields struct {
		ID   string
		Size int64
	}
	type args struct {
		finishedChunks []Chunk
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantFinished bool
	}{
		{
			name:   "simple complete",
			fields: fields{ID: "o1", Size: 1000},
			args: args{finishedChunks: []Chunk{
				{Size: 500, Offset: 500, Object: Object{ID: "o1", Size: 1000}},
				{Size: 500, Offset: 0, Object: Object{ID: "o1", Size: 1000}},
			}},
			wantFinished: true,
		},
		{
			name:   "simple incomplete",
			fields: fields{ID: "o1", Size: 1000},
			args: args{finishedChunks: []Chunk{
				{Size: 450, Offset: 500, Object: Object{ID: "o1", Size: 1000}},
				{Size: 500, Offset: 0, Object: Object{ID: "o1", Size: 1000}},
			}},
			wantFinished: false,
		},
		{
			name:   "overlapping complete",
			fields: fields{ID: "o1", Size: 1000},
			args: args{finishedChunks: []Chunk{
				{Size: 200, Offset: 100, Object: Object{ID: "o1", Size: 1000}},
				{Size: 950, Offset: 50, Object: Object{ID: "o1", Size: 1000}},
				{Size: 200, Offset: 0, Object: Object{ID: "o1", Size: 1000}},
			}},
			wantFinished: true,
		},
		{
			name:   "overlapping incomplete",
			fields: fields{ID: "o1", Size: 1000},
			args: args{finishedChunks: []Chunk{
				{Size: 200, Offset: 100, Object: Object{ID: "o1", Size: 1000}},
				{Size: 850, Offset: 50, Object: Object{ID: "o1", Size: 1000}},
				{Size: 200, Offset: 0, Object: Object{ID: "o1", Size: 1000}},
			}},
			wantFinished: false,
		},
		{
			name:   "overlapping complete with duplicated offsets",
			fields: fields{ID: "o1", Size: 1000},
			args: args{finishedChunks: []Chunk{
				{Size: 200, Offset: 100, Object: Object{ID: "o1", Size: 1000}},
				{Size: 300, Offset: 50, Object: Object{ID: "o1", Size: 1000}},
				{Size: 900, Offset: 50, Object: Object{ID: "o1", Size: 1000}},
				{Size: 500, Offset: 50, Object: Object{ID: "o1", Size: 1000}},
				{Size: 200, Offset: 0, Object: Object{ID: "o1", Size: 1000}},
				{Size: 100, Offset: 900, Object: Object{ID: "o1", Size: 1000}},
			}},
			wantFinished: true,
		},
		{
			name:   "trailing edge incomplete",
			fields: fields{ID: "o1", Size: 1000},
			args: args{finishedChunks: []Chunk{
				{Size: 499, Offset: 500, Object: Object{ID: "o1", Size: 1000}},
				{Size: 500, Offset: 0, Object: Object{ID: "o1", Size: 1000}},
			}},
			wantFinished: false,
		},
		{
			name:   "ignore chunks from different objects",
			fields: fields{ID: "o1", Size: 1000},
			args: args{finishedChunks: []Chunk{
				{Size: 500, Offset: 500, Object: Object{ID: "o2", Size: 1000}},
				{Size: 500, Offset: 0, Object: Object{ID: "o1", Size: 1000}},
			}},
			wantFinished: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := Object{
				ID:   tt.fields.ID,
				Size: tt.fields.Size,
			}
			if gotFinished := o.isCompletedBy(tt.args.finishedChunks); gotFinished != tt.wantFinished {
				t.Errorf("Object.isCompletedBy() = %v, want %v", gotFinished, tt.wantFinished)
			}
		})
	}
}
