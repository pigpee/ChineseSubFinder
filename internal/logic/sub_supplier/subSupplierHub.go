package sub_supplier

import (
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/ifaces"
	movieHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/movie_helper"
	seriesHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/series_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/sirupsen/logrus"
	"gopkg.in/errgo.v2/fmt/errors"
	"path/filepath"
)

type SubSupplierHub struct {
	settings settings.Settings

	Suppliers []ifaces.ISupplier

	log *logrus.Logger
}

func NewSubSupplierHub(_settings settings.Settings, one ifaces.ISupplier, _inSupplier ...ifaces.ISupplier) *SubSupplierHub {
	s := SubSupplierHub{}
	s.settings = _settings
	s.log = log_helper.GetLogger()
	s.Suppliers = make([]ifaces.ISupplier, 0)
	s.Suppliers = append(s.Suppliers, one)
	if len(_inSupplier) > 0 {
		for _, supplier := range _inSupplier {
			s.Suppliers = append(s.Suppliers, supplier)
		}
	}

	return &s
}

// AddSubSupplier 添加一个下载器，目前目标是给 SubHD 使用
func (d *SubSupplierHub) AddSubSupplier(one ifaces.ISupplier) {
	d.Suppliers = append(d.Suppliers, one)
}

func (d *SubSupplierHub) DelSubSupplier(one ifaces.ISupplier) {

	for i := 0; i < len(d.Suppliers); i++ {

		if one.GetSupplierName() == d.Suppliers[i].GetSupplierName() {
			d.Suppliers = append(d.Suppliers[:i], d.Suppliers[i+1:]...)
		}
	}
}

// DownloadSub4Movie 某一个电影字幕下载，下载完毕后，返回下载缓存每个字幕的位置
func (d SubSupplierHub) DownloadSub4Movie(videoFullPath string, index int, forcedScanAndDownloadSub bool) ([]string, error) {

	if forcedScanAndDownloadSub == false {
		// 非强制扫描的时候，需要判断这个视频根目录是否有 .ignore 文件，有也跳过
		if my_util.IsFile(filepath.Join(filepath.Dir(videoFullPath), common.Ignore)) == true {
			d.log.Infoln("Found", common.Ignore, "Skip", videoFullPath)
			// 跳过下载字幕
			return nil, nil
		}
	}

	// 跳过中文的电影，不是一定要跳过的
	skip, err := movieHelper.SkipChineseMovie(videoFullPath, *d.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		d.log.Warnln("SkipChineseMovie", videoFullPath, err)
	}
	if skip == true {
		return nil, nil
	}
	var needDlSub = false
	if forcedScanAndDownloadSub == true {
		// 强制下载字幕
		needDlSub = true
	} else {
		needDlSub, err = movieHelper.MovieNeedDlSub(videoFullPath)
		if err != nil {
			return nil, errors.Newf("MovieNeedDlSub %v %v", videoFullPath, err)
		}
	}
	if needDlSub == true {
		// 需要下载字幕
		// 下载所有字幕
		subInfos := movieHelper.OneMovieDlSubInAllSite(d.Suppliers, videoFullPath, index)
		// 整理字幕，比如解压什么的
		organizeSubFiles, err := sub_helper.OrganizeDlSubFiles(filepath.Base(videoFullPath), subInfos)
		if err != nil {
			return nil, errors.Newf("OrganizeDlSubFiles %v %v", videoFullPath, err)
		}
		// 因为是下载电影，需要合并返回
		var outSubFileFullPathList = make([]string, 0)
		for s := range organizeSubFiles {
			outSubFileFullPathList = append(outSubFileFullPathList, organizeSubFiles[s]...)
		}

		for i, subFile := range outSubFileFullPathList {
			d.log.Debugln("OneMovieDlSubInAllSite", videoFullPath, i, "SubFileFPath:", subFile)
		}

		return outSubFileFullPathList, nil
	} else {
		// 无需下载字幕
		return nil, nil
	}
}

// DownloadSub4Series 某一部连续剧的字幕下载，下载完毕后，返回下载缓存每个字幕的位置
func (d SubSupplierHub) DownloadSub4Series(seriesDirPath string, index int, forcedScanAndDownloadSub bool) (*series.SeriesInfo, map[string][]string, error) {

	if forcedScanAndDownloadSub == false {
		// 非强制扫描的时候，需要判断这个视频根目录是否有 .ignore 文件，有也跳过
		if my_util.IsFile(filepath.Join(seriesDirPath, common.Ignore)) == true {
			d.log.Infoln("Found", common.Ignore, "Skip", seriesDirPath)
			// 跳过下载字幕
			return nil, nil, nil
		}
	}

	// 跳过中文的连续剧，不是一定要跳过的
	skip, imdbInfo, err := seriesHelper.SkipChineseSeries(seriesDirPath, *d.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		d.log.Warnln("SkipChineseSeries", seriesDirPath, err)
	}
	if skip == true {
		return nil, nil, nil
	}
	// 读取本地的视频和字幕信息
	seriesInfo, err := seriesHelper.ReadSeriesInfoFromDir(seriesDirPath, imdbInfo, forcedScanAndDownloadSub)
	if err != nil {
		return nil, nil, errors.Newf("ReadSeriesInfoFromDir %v %v", seriesDirPath, err)
	}
	organizeSubFiles, err := d.dlSubFromSeriesInfo(seriesDirPath, index, seriesInfo, err)
	if err != nil {
		return nil, nil, err
	}
	return seriesInfo, organizeSubFiles, nil
}

// DownloadSub4SeriesFromEmby 通过 Emby 查询到的信息进行字幕下载，下载完毕后，返回下载缓存每个字幕的位置
func (d SubSupplierHub) DownloadSub4SeriesFromEmby(seriesDirPath string, seriesList []emby.EmbyMixInfo, index int) (*series.SeriesInfo, map[string][]string, error) {

	// 跳过中文的连续剧，不是一定要跳过的
	skip, imdbInfo, err := seriesHelper.SkipChineseSeries(seriesDirPath, *d.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		d.log.Warnln("SkipChineseSeries", seriesDirPath, err)
	}
	if skip == true {
		return nil, nil, nil
	}
	// 读取本地的视频和字幕信息
	seriesInfo, err := seriesHelper.ReadSeriesInfoFromEmby(seriesDirPath, imdbInfo, seriesList)
	if err != nil {
		return nil, nil, errors.Newf("ReadSeriesInfoFromDir %v %v", seriesDirPath, err)
	}
	organizeSubFiles, err := d.dlSubFromSeriesInfo(seriesDirPath, index, seriesInfo, err)
	if err != nil {
		return nil, nil, err
	}
	return seriesInfo, organizeSubFiles, nil
}

// CheckSubSiteStatus 检测多个字幕提供的网站是否是有效的
func (d *SubSupplierHub) CheckSubSiteStatus() backend.ReplyCheckStatus {

	outStatus := backend.ReplyCheckStatus{
		SubSiteStatus: make([]backend.SiteStatus, 0),
	}

	// 测试提供字幕的网站是有效的
	d.log.Infoln("Check Sub Supplier Start...")
	for _, supplier := range d.Suppliers {
		bAlive, speed := supplier.CheckAlive()
		if bAlive == false {
			d.log.Warningln(supplier.GetSupplierName(), "Check Alive = false")
		} else {
			d.log.Infoln(supplier.GetSupplierName(), "Check Alive = true, Speed =", speed, "ms")
		}

		outStatus.SubSiteStatus = append(outStatus.SubSiteStatus, backend.SiteStatus{
			Name:  supplier.GetSupplierName(),
			Valid: bAlive,
			Speed: speed,
		})
	}

	suppliersLen := len(d.Suppliers)
	for i := 0; i < suppliersLen; i++ {
		if d.Suppliers[i].IsAlive() == false {
			d.DelSubSupplier(d.Suppliers[i])
		}
		suppliersLen = len(d.Suppliers)
	}

	d.log.Infoln("Check Sub Supplier End")

	return outStatus
}

func (d SubSupplierHub) dlSubFromSeriesInfo(seriesDirPath string, index int, seriesInfo *series.SeriesInfo, err error) (map[string][]string, error) {
	// 下载好的字幕
	subInfos := seriesHelper.DownloadSubtitleInAllSiteByOneSeries(d.Suppliers, seriesInfo, index)
	// 整理字幕，比如解压什么的
	// 每一集 SxEx - 对应解压整理后的字幕列表
	organizeSubFiles, err := sub_helper.OrganizeDlSubFiles(filepath.Base(seriesDirPath), subInfos)
	if err != nil {
		return nil, errors.Newf("OrganizeDlSubFiles %v %v", seriesDirPath, err)
	}
	return organizeSubFiles, nil
}
