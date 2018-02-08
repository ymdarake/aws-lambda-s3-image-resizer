package main

/**
 * json形式のルールデータを返します。
 * pathは正規表現として解釈されます。
 */
func RawRules() []byte {
	return []byte(`
[
	{
		"path": "test-test/test/[a-zA-Z0-9]+\\.original\\.(jpg|jpeg|gif|png)",
		"outputspecs": [
			{
				"x": 120,				
				"y": 120,				
				"directory": "#ORIG_DIR"
			},
			{
				"x": 600,
				"y": 600,
				"directory": "#ORIG_DIR"
			}
		]
	},
	{
		"path": "test-test/specific-image-file\\.original\\.(jpg|jpeg|gif|png)",
		"outputspecs": [
			{
				"x": 300,				
				"y": 200,				
				"directory": "#ORIG_DIR"
			},
			{
				"x": 480,
				"y": 480,
				"directory": "#ORIG_DIR"
			},
			{
				"x": 1200,
				"y": 1200,
				"directory": "specific-dir/specific-name.ext"
			}
		]
	}
]`)
}
