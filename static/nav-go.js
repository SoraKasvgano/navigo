// nav-go.js - Navigation page JavaScript for Go backend
// Based on nav.js, adapted for goversion independent deployment

function getQueryVariable(variable) {
    var query = window.location.search.substring(1);
    var vars = query.split("&");
    for (var i = 0; i < vars.length; i++) {
        var pair = vars[i].split("=");
        if (pair[0] == variable) {
            return pair[1];
        }
    }
    return (false);
}

function replaceParamVal(url, paramName, replaceVal) {
    var oUrl = url.toString();
    var re = eval('/(' + paramName + '=)([^&]*)/gi');
    var nUrl = oUrl.replace(re, paramName + '=' + replaceVal);
    return nUrl;
}

function setCookie(cname, cvalue, exsecond) {
    var d = new Date();
    d.setTime(d.getTime() + (exsecond * 1000));
    var expires = "expires=" + d.toUTCString();
    document.cookie = cname + "=" + cvalue + ";" + expires + ";path=/";
}

function getCookie(name) {
    var arr, reg = new RegExp("(^| )" + name + "=([^;]*)(;|$)");
    if (arr = document.cookie.match(reg))
        return unescape(arr[2]);
    else
        return null;
}

function getArrayIndex(arr, obj) {
    var i = arr.length;
    while (i--) {
        if (arr[i] === obj) {
            return i;
        }
    }
    return -1;
}

var f_Array = "";

$(function() {
    // Load nav.json from runtime-generated file
    var navUrl = "/nav.json?t=" + new Date().getTime();

    $.getJSON(navUrl, function(data) {
        // Handle page config
        var pageConfig = data.find(function(item) {
            return item.type === "page_config";
        });

        if (pageConfig) {
            // Update page title
            if (pageConfig.title) {
                document.title = pageConfig.title;
                $('.left-bar .title p').text(pageConfig.title);
            }

            // Update subtitle
            if (pageConfig.subtitle) {
                $('.left-bar .ti-sec p').text(pageConfig.subtitle);
            }

            // Update logo
            if (pageConfig.logo) {
                $('.left-bar .big-logo img').attr('src', pageConfig.logo);
            }

            // Update footer text
            if (pageConfig.footer_text) {
                $('.footer .copyright .copyright-text p').text(pageConfig.footer_text);
            }

            // Update ICP
            if (pageConfig.icp) {
                $('.left-bar .fixed-bottom .icp').html('<a href="https://beian.miit.gov.cn/" target="_blank">' + pageConfig.icp + '</a>');
            }

            // Remove page config from data array
            data = data.filter(function(item) {
                return item.type !== "page_config";
            });
        }

        // Handle announcement config
        var announcementConfig = data.find(function(item) {
            return item.type === "announcement_config";
        });

        if (announcementConfig && announcementConfig.announcements && announcementConfig.announcements.length > 0) {
            initAnnouncementScroll(announcementConfig);
            data = data.filter(function(item) {
                return item.type !== "announcement_config";
            });
        }

        var strHtml = "";
        $.each(data, function(infoIndex, info) {
            var navstr = "";
            var navtitle = "";
            strHtml += "<li><a href='#" + info["_id"] + "'><span class ='" + info["icon"] + "'></span>" + info["classify"] + "</a></li>";
            navtitle += "<div class='box box_default'><a href='#' id='" + info["_id"] + "'></a> <div class='sub-category'> <div><span class='" + info["icon"] + "'></span>" + info["classify"] + "</div> </div><div>";
            $.each(info["sites"], function(i, str) {
                if (str["logo"] == "no-logo") {
                    str["logo"] = "/static/logo.svg";
                }
                navstr += '<a target="_blank" href="' + str["href"] + '">';
                navstr += '<div class="item">';
                navstr += '    <div class="logo">'
                navstr += '       <img src="' + str["logo"] + '"></div> ';
                navstr += '   <div class="content">' + '<strong>' + str["name"] + '</strong><p class="desc">' + str["desc"] + '</p></div>';
                navstr += '</div>      </a>';
            })
            navstr = navtitle + navstr + '</div>';
            $(".footer").before(navstr);
        })
        $("#navItem").append(strHtml);
    });

    // Custom module rendering
    if (getQueryVariable('p') != "") {
        $.getJSON(getQueryVariable('p'), function(data) {
            var f = getQueryVariable('f')
            fArray = data.map(function(item) {
                return item.filter;
            });
            if (!Array.from) {
                Array.from = function(el) {
                    return Array.apply(this, el);
                }
            }
            f_Array = Array.from(new Set(fArray));
            var strHtml = "";
            if (f != "") {
                var model = data.filter(function(e) {
                    return e.filter == f;
                });
                $.each(model, function(infoIndex, info) {
                    strHtml += "<li><a href='#" + info["_id"] + "'><span class ='" + info["icon"] + "'></span>" + info["classify"] + "[Custom]</a></li>";
                })
                $.each(model.reverse(), function(infoIndex, info) {
                    var navstr = "";
                    var navtitle = "";
                    navtitle += "<div class='box box_user'><a href='#' id='" + info["_id"] + "'></a> <div class='sub-category'> <div><i class='" + info["icon"] + "'></i>" + info["classify"] + "[Custom]</div> </div><div>";
                    $.each(info["sites"], function(i, str) {
                        if (str["logo"] == "no-logo") {
                            str["logo"] = "/static/logo.svg";
                        }
                        navstr += '<a target="_blank" href="' + str["href"] + '">';
                        navstr += '<div class="item">';
                        navstr += '    <div class="logo">'
                        navstr += '       <img src="' + str["logo"] + '"></div> ';
                        navstr += '   <div class="content_d">' + '<strong>' + str["name"] + '</strong></div>';
                        navstr += '</div>      </a>';
                    })
                    navstr = navtitle + navstr + '</div>';
                    $(".about").after(navstr);
                })

                // Quick refresh filter switching
                last_time = getCookie("time@" + f);
                if (last_time != "" && last_time != null) {
                    setCookie("time@" + f, "", -1);
                    href = window.location.href;
                    var findex = getArrayIndex(f_Array, f);
                    var indexto = findex + 1;
                    if (indexto > f_Array.length - 1) {
                        indexto = 0;
                    }
                    self.location.href = replaceParamVal(href, "f", f_Array[indexto]);
                }
                d = new Date();
                ctime = d.getSeconds();
                setCookie("time@" + f, ctime.toString(), 3);
            } else {
                $.each(data, function(infoIndex, info) {
                    strHtml += "<li><a href='#" + info["_id"] + "'><span class ='" + info["icon"] + "'></span>" + info["classify"] + "[Custom]</a></li>";
                })
                $.each(data.reverse(), function(infoIndex, info) {
                    var navstr = "";
                    var navtitle = "";
                    navtitle += "<div class='box box_user'><a href='#' id='" + info["_id"] + "'></a> <div class='sub-category'> <div><i class='" + info["icon"] + "'></i>" + info["classify"] + "[Custom]</div> </div><div>";
                    $.each(info["sites"], function(i, str) {
                        if (str["logo"] == "no-logo") {
                            str["logo"] = "/static/logo.svg";
                        }
                        navstr += '<a target="_blank" href="' + str["href"] + '">';
                        navstr += '<div class="item">';
                        navstr += '    <div class="logo">'
                        navstr += '       <img src="' + str["logo"] + '"></div> ';
                        navstr += '   <div class="content_d">' + '<strong>' + str["name"] + '</strong></div>';
                        navstr += '</div>      </a>';
                    })
                    navstr = navtitle + navstr + '</div>';
                    $(".about").after(navstr);
                })
            }

            $("#navItem").prepend(strHtml);
        })
    }

    // Navigation bar click positioning
    var href = "";
    var pos = "";
    $(".nav-item ").on('click', 'li', function(e) {
        $(".nav-item li a").each(function() {
            $(this).removeClass("active");
        });

        $(this).children().addClass("active");

        e.preventDefault();
        href = $(this).children().attr("href");
        pos = $(href).position().top - 30;
        $("html,body").animate({
            scrollTop: pos
        }, 500);
    });

    // Mobile navigation bar toggle
    var oMenu = document.getElementById('menu');
    var oBtn = oMenu.getElementsByTagName('a')[0];
    var oLeftBar = document.getElementById('leftBar');
    oBtn.onclick = function() {
        if (oLeftBar.offsetLeft == 0) {
            oLeftBar.style.left = -249 + 'px';
        } else {
            oLeftBar.style.left = 0 + 'px';
        }
        if (document.documentElement.clientWidth <= 600) {
            document.onclick = function() {
                if (oLeftBar.offsetLeft == 0) {
                    oLeftBar.style.left = -249 + 'px';
                }
            }
        }
    }

    // Scroll to top
    $(window).scroll(function() {
        if ($(window).scrollTop() >= 200) {
            $('#fixedBar').fadeIn(300);
        } else {
            $('#fixedBar').fadeOut(300);
        }
    });
    $('#fixedBar').click(function() {
        $('html,body').animate({
            scrollTop: '0px'
        }, 800);
    });

    // Initialize calendar
    initCalendar();

    // Initialize world time
    setTimeout(initWorldTime, 100);
});

// Update calendar
function updateCalendar() {
    try {
        const now = new Date();
        const lunar = Lunar.fromDate(now);
        const solar = lunar.getSolar();

        const weekMap = {
            '0': 'Sun', '1': 'Mon', '2': 'Tue', '3': 'Wed',
            '4': 'Thu', '5': 'Fri', '6': 'Sat'
        };
        const weekDay = weekMap[solar.getWeek()] || 'Week ' + solar.getWeek();
        const dateTimeStr = `${solar.getYear()}-${String(solar.getMonth()).padStart(2, '0')}-${String(solar.getDay()).padStart(2, '0')} ${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')} ${weekDay} ${solar.getXingZuo()}`;

        // Update basic info
        document.querySelector('.lunar-date').textContent = lunar.toString();
        document.querySelector('.lunar-week').textContent = weekDay;
        document.querySelector('.lunar-constellation').textContent = solar.getXingZuo();

        // Update Ganzhi and Nayin info
        document.querySelector('.year-ganzhi').textContent = `${lunar.getYearInGanZhi()}(${lunar.getYearShengXiao()})`;
        document.querySelector('.month-ganzhi').textContent = `${lunar.getMonthInGanZhi()}(${lunar.getMonthShengXiao()})`;
        document.querySelector('.day-ganzhi').textContent = `${lunar.getDayInGanZhi()}(${lunar.getDayShengXiao()})`;
        document.querySelector('.time-ganzhi').textContent = `${lunar.getTimeZhi()}(${lunar.getTimeShengXiao()})`;
        document.querySelector('.nayin-content').textContent = `${lunar.getYearNaYin()} ${lunar.getMonthNaYin()} ${lunar.getDayNaYin()} ${lunar.getTimeNaYin()}`;

        // Update Xingsu and Shensha info
        document.querySelector('.xingsu-content').textContent = `${lunar.getXiu()}${lunar.getZheng()}${lunar.getAnimal()}(${lunar.getXiuLuck()})`;
        document.querySelector('.pengzu-content').textContent = `${lunar.getPengZuGan()} ${lunar.getPengZuZhi()}`;
        document.querySelector('.jishen-content').textContent = lunar.getDayJiShen();
        document.querySelector('.xiongsha-content').textContent = lunar.getDayXiongSha();

        // Update direction and taboo info
        document.querySelector('.xishen-content').textContent = `${lunar.getDayPositionXi()}(${lunar.getDayPositionXiDesc()})`;
        document.querySelector('.yangguishen-content').textContent = `${lunar.getDayPositionYangGui()}(${lunar.getDayPositionYangGuiDesc()})`;
        document.querySelector('.yinguishen-content').textContent = `${lunar.getDayPositionYinGui()}(${lunar.getDayPositionYinGuiDesc()})`;
        document.querySelector('.fushen-content').textContent = `${lunar.getDayPositionFu()}(${lunar.getDayPositionFuDesc()})`;
        document.querySelector('.caishen-content').textContent = `${lunar.getDayPositionCai()}(${lunar.getDayPositionCaiDesc()})`;
        document.querySelector('.chong-content').textContent = lunar.getDayChongDesc();
        document.querySelector('.sha-content').textContent = lunar.getDaySha();

        // Update DOM
        document.querySelector('.date-time').textContent = dateTimeStr;
    } catch (error) {
        console.error('Failed to update calendar:', error);
        document.querySelector('.date-time').textContent = 'Loading failed';
    }
}

// Initialize calendar
function initCalendar() {
    updateCalendar();
    setInterval(updateCalendar, 1000);
    initLunarCalendarToggle();
}

// Lunar calendar toggle
function initLunarCalendarToggle() {
    const toggleBtn = document.querySelector('.toggle-btn');
    const lunarCalendar = document.querySelector('.lunar-calendar');
    const detailContent = document.querySelector('.lunar-detail-content');

    if (!toggleBtn || !lunarCalendar || !detailContent) {
        return;
    }

    function expandDetails() {
        detailContent.style.display = 'block';
        setTimeout(() => {
            lunarCalendar.classList.add('expanded');
            toggleBtn.classList.add('expanded');
            detailContent.classList.add('show');
        }, 0);
    }

    function collapseDetails() {
        lunarCalendar.classList.remove('expanded');
        toggleBtn.classList.remove('expanded');
        detailContent.classList.remove('show');
        setTimeout(() => {
            detailContent.style.display = 'none';
        }, 300);
    }

    toggleBtn.addEventListener('click', function(event) {
        event.stopPropagation();
        const isExpanded = lunarCalendar.classList.contains('expanded');

        if (isExpanded) {
            collapseDetails();
        } else {
            expandDetails();
        }
    });

    lunarCalendar.addEventListener('click', function(event) {
        event.stopPropagation();
    });

    document.addEventListener('click', function() {
        if (lunarCalendar.classList.contains('expanded')) {
            collapseDetails();
        }
    });
}

// World time
function initWorldTime() {
    updateWorldTime();
    setInterval(updateWorldTime, 1000);
}

function updateWorldTime() {
    try {
        var now = new Date();
        var localOffset = now.getTimezoneOffset() * 60 * 1000;

        // Beijing (UTC+8)
        var beijingTime = new Date(now.getTime() + localOffset + (8 * 60 * 60 * 1000));
        document.getElementById('beijingTime').innerHTML = formatDate(beijingTime) + '<br>' + formatTime(beijingTime);

        // London (UTC+0)
        var londonTime = new Date(now.getTime() + localOffset);
        document.getElementById('londonTime').innerHTML = formatDate(londonTime) + '<br>' + formatTime(londonTime);

        // Amsterdam (UTC+1)
        var amsterdamTime = new Date(now.getTime() + localOffset + (1 * 60 * 60 * 1000));
        document.getElementById('amsterdamTime').innerHTML = formatDate(amsterdamTime) + '<br>' + formatTime(amsterdamTime);

        // New York (UTC-5)
        var newyorkTime = new Date(now.getTime() + localOffset - (5 * 60 * 60 * 1000));
        document.getElementById('newyorkTime').innerHTML = formatDate(newyorkTime) + '<br>' + formatTime(newyorkTime);

        // Sydney (UTC+10)
        var sydneyTime = new Date(now.getTime() + localOffset + (10 * 60 * 60 * 1000));
        document.getElementById('sydneyTime').innerHTML = formatDate(sydneyTime) + '<br>' + formatTime(sydneyTime);

    } catch (error) {
        console.error('Failed to update world time:', error);
    }
}

function formatTime(date) {
    return date.getHours().toString().padStart(2, '0') + ':' +
        date.getMinutes().toString().padStart(2, '0');
}

function formatDate(date) {
    var month = date.getMonth() + 1;
    var day = date.getDate();
    var weekdays = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
    var weekday = weekdays[date.getDay()];
    return month + '/' + day + ' ' + weekday;
}

// Announcement scroll
function initAnnouncementScroll(config) {
    'use strict';

    var announcements = config.announcements || [];
    var refreshInterval = config.interval || 5000;
    var currentIndex = 0;

    if (announcements.length === 0) {
        return;
    }

    var aboutBox = document.querySelector('.about');
    if (!aboutBox) {
        return;
    }

    // Modify about box HTML to support announcement scroll and search
    aboutBox.innerHTML = `
        <div class="about-content" style="flex: 4;">
            <div class="announcement-container" style="position: relative; height: 35px; overflow: hidden;">
                <div class="announcement-current" style="height: 35px; display: flex; align-items: center; padding: 0 15px;">
                    <i class="ti-pencil-alt" style="color: #007bff; font-size: 12px;"></i>
                    <span class="announcement-time" style="font-size: 11px; color: #6c757d; font-weight: bold; margin-right: 8px; flex-shrink: 0; min-width: 50px;">${formatTimestamp(announcements[0].timestamp)}</span>
                    <span class="announcement-text" style="flex: 1; color: #495057; font-size: 13px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">${announcements[0].content}</span>
                </div>
            </div>
            ${announcements.length > 1 ? `
            <div class="announcement-indicators" style="position: absolute; bottom: 5px; left: 0; width: 100%; display: flex; justify-content: center; gap: 6px;">
                ${announcements.map(function(announcement, index) {
                    return `<div class="announcement-indicator" data-index="${index}" style="width: ${index === 0 ? '12px' : '6px'}; height: 6px; border-radius: ${index === 0 ? '3px' : '50%'}; background: ${index === 0 ? '#007bff' : '#ccc'}; cursor: pointer; transition: all 0.3s ease;"></div>`;
                }).join('')}
            </div>
            ` : ''}
        </div>
        <!-- Search box -->
        <div class="searchForm">
            <div class="searchBarAc">
                <div class="oicon"><i class="ti-search"></i></div>
                <div class="textInput">
                    <input type="text" placeholder="Search..." id="searchInput">
                    <div class="clear"><i class="ti-close"></i></div>
                    <div class="search-options">
                        <select id="searchType">
                            <option value="bing" selected>Bing</option>
                            <option value="local">Local</option>
                        </select>
                    </div>
                    <button id="searchBtn">Search</button>
                </div>
            </div>
        </div>
        <!-- Search results overlay -->
        <div class="search-results-overlay" id="searchResultsOverlay" style="display: none;"></div>

        <!-- Search results frame -->
        <div class="search-results" id="searchResults" style="display: none;">
            <div class="search-results-header">
                <h3>Search Results</h3>
                <button class="back-btn" id="backBtn">
                    <i class="ti-arrow-left"></i> Back
                </button>
            </div>
            <div class="search-results-content" id="searchResultsContent">
            </div>
        </div>
    `;

    var announcementCurrent = aboutBox.querySelector('.announcement-current');
    var indicatorsContainer = aboutBox.querySelector('.announcement-indicators');
    var indicators = aboutBox.querySelectorAll('.announcement-indicator');

    function updateCurrentAnnouncement(index) {
        if (index < 0 || index >= announcements.length) return;

        currentIndex = index;
        var announcement = announcements[index];
        var timestamp = formatTimestamp(announcement.timestamp);

        announcementCurrent.style.opacity = '0';
        setTimeout(function() {
            announcementCurrent.innerHTML = `
                <i class="ti-pencil-alt" style="color: #007bff; font-size: 12px;"></i>
                <span class="announcement-time" style="font-size: 11px; color: #6c757d; font-weight: bold; margin-right: 8px; flex-shrink: 0; min-width: 50px;">${timestamp}</span>
                <span class="announcement-text" style="flex: 1; color: #495057; font-size: 13px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">${announcement.content}</span>
            `;
            announcementCurrent.style.opacity = '1';
        }, 200);

        updateIndicators();
    }

    function updateIndicators() {
        if (indicators.length > 0) {
            indicators.forEach(function(indicator, index) {
                if (index === currentIndex) {
                    indicator.style.background = '#007bff';
                    indicator.style.width = '12px';
                    indicator.style.borderRadius = '3px';
                } else {
                    indicator.style.background = '#ccc';
                    indicator.style.width = '6px';
                    indicator.style.borderRadius = '50%';
                }
            });
        }
    }

    function refreshAnnouncement() {
        if (announcements.length <= 1) {
            return;
        }

        var nextIndex = (currentIndex + 1) % announcements.length;
        updateCurrentAnnouncement(nextIndex);
    }

    var refreshTimer = setInterval(refreshAnnouncement, refreshInterval);

    var announcementContainer = aboutBox.querySelector('.announcement-container');
    if (announcementContainer) {
        announcementContainer.addEventListener('mouseenter', function() {
            clearInterval(refreshTimer);
        });

        announcementContainer.addEventListener('mouseleave', function() {
            refreshTimer = setInterval(refreshAnnouncement, refreshInterval);
        });
    }

    if (indicators.length > 0) {
        indicators.forEach(function(indicator, index) {
            indicator.addEventListener('click', function() {
                clearInterval(refreshTimer);
                updateCurrentAnnouncement(index);
                refreshTimer = setInterval(refreshAnnouncement, refreshInterval);
            });
        });
    }

    function toggleIndicators() {
        if (indicatorsContainer) {
            indicatorsContainer.style.display = 'flex';
        }
    }

    toggleIndicators();
    window.addEventListener('resize', toggleIndicators);

    setTimeout(initSearchBox, 0);
}

// Format timestamp
function formatTimestamp(timestamp) {
    try {
        var date = new Date(timestamp);
        if (isNaN(date.getTime())) {
            return timestamp;
        }

        var now = new Date();
        var diff = now - date;
        var diffDays = Math.floor(diff / (1000 * 60 * 60 * 24));

        if (diffDays === 0) {
            return 'Today ' + date.toLocaleTimeString('en-US', {
                hour: '2-digit',
                minute: '2-digit'
            });
        } else if (diffDays === 1) {
            return 'Yesterday ' + date.toLocaleTimeString('en-US', {
                hour: '2-digit',
                minute: '2-digit'
            });
        } else if (diffDays < 7) {
            return diffDays + ' days ago';
        } else {
            return date.toLocaleDateString('en-US', {
                month: '2-digit',
                day: '2-digit'
            });
        }
    } catch (error) {
        return timestamp;
    }
}

// Search box
function initSearchBox() {
    var searchInput = document.getElementById('searchInput');
    var searchBtn = document.getElementById('searchBtn');
    var clearBtn = document.querySelector('.clear');
    var searchType = document.getElementById('searchType');
    var searchResults = document.getElementById('searchResults');
    var searchResultsContent = document.getElementById('searchResultsContent');
    var backBtn = document.getElementById('backBtn');
    var searchResultsOverlay = document.getElementById('searchResultsOverlay');

    if (!searchInput || !searchBtn || !clearBtn || !searchType || !searchResults || !searchResultsContent || !backBtn || !searchResultsOverlay) {
        setTimeout(initSearchBox, 1000);
        return;
    }

    clearBtn.addEventListener('click', function() {
        searchInput.value = '';
        searchInput.focus();
        updateClearButtonVisibility();
        hideSearchResults();
    });

    searchInput.addEventListener('input', function() {
        updateClearButtonVisibility();
        if (searchInput.value.trim() === '') {
            hideSearchResults();
        }
    });

    searchBtn.addEventListener('click', performSearch);
    searchInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            performSearch();
        }
    });

    backBtn.addEventListener('click', hideSearchResults);

    updateClearButtonVisibility();

    function updateClearButtonVisibility() {
        if (searchInput.value.trim() !== '') {
            clearBtn.style.opacity = '1';
        } else {
            clearBtn.style.opacity = '0';
        }
    }

    function performSearch() {
        var query = searchInput.value.trim();
        if (query === '') {
            searchInput.focus();
            return;
        }

        var selectedType = searchType.value;

        if (selectedType === 'local') {
            performLocalSearch(query);
        } else if (selectedType === 'bing') {
            performBingSearch(query);
        }
    }

    function performLocalSearch(query) {
        var navUrl = "/nav.json?t=" + new Date().getTime();

        $.getJSON(navUrl, function(data) {
            var searchData = data.filter(function(item) {
                return item.type !== "announcement_config";
            });

            var results = searchLocalData(searchData, query);
            displaySearchResults(results, query);
        }).fail(function() {
            console.error('Failed to load nav.json');
            alert('Failed to load search data, please refresh and try again');
        });
    }

    function searchLocalData(data, query) {
        var results = [];
        var lowerQuery = query.toLowerCase();

        data.forEach(function(category) {
            if (category.classify && category.classify.toLowerCase().includes(lowerQuery)) {
                results.push({
                    type: 'category',
                    category: category.classify,
                    icon: category.icon,
                    match: category.classify
                });
            }

            if (category.sites && Array.isArray(category.sites)) {
                category.sites.forEach(function(site) {
                    var matchFound = false;
                    var matchText = '';

                    if (site.name && site.name.toLowerCase().includes(lowerQuery)) {
                        matchFound = true;
                        matchText = site.name;
                    } else if (site.desc && site.desc.toLowerCase().includes(lowerQuery)) {
                        matchFound = true;
                        matchText = site.desc;
                    } else if (site.href && site.href.toLowerCase().includes(lowerQuery)) {
                        matchFound = true;
                        matchText = site.href;
                    }

                    if (matchFound) {
                        results.push({
                            type: 'site',
                            category: category.classify,
                            name: site.name,
                            desc: site.desc,
                            href: site.href,
                            logo: site.logo,
                            match: matchText
                        });
                    }
                });
            }
        });

        return results;
    }

    function displaySearchResults(results, query) {
        searchResultsOverlay.style.display = 'block';
        searchResults.style.display = 'block';
        searchResults.classList.add('box');

        var html = '';

        if (results.length === 0) {
            html = '<div class="no-results">No results found for "' + query + '"</div>';
        } else {
            html = '<div class="search-summary" style="margin-bottom: 20px; font-size: 14px; color: #666;">Found ' + results.length + ' results for "' + query + '"</div>';

            var categories = {};

            results.forEach(function(result) {
                if (!categories[result.category]) {
                    categories[result.category] = [];
                }
                categories[result.category].push(result);
            });

            Object.keys(categories).forEach(function(categoryName) {
                html += '<div class="search-category" style="margin-bottom: 25px;">';
                html += '<h4 class="category-title" style="font-size: 16px; color: #333; margin-bottom: 15px; border-bottom: 1px solid #e9ecef; padding-bottom: 8px;">' + categoryName + '</h4>';

                categories[categoryName].forEach(function(result) {
                    if (result.type === 'category') {
                        html += '<a target="_blank" href="javascript:void(0)">';
                        html += '<div class="search-result-item">';
                        html += '<div class="logo">';
                        html += '<i class="' + result.icon + '" style="font-size: 24px; color: #3273dc;"></i>';
                        html += '</div>';
                        html += '<div class="content">';
                        html += '<strong>Category: ' + result.match + '</strong>';
                        html += '<p class="desc">Click to view all sites in this category</p>';
                        html += '</div>';
                        html += '</div>';
                        html += '</a>';
                    } else if (result.type === 'site') {
                        html += '<a target="_blank" href="' + result.href + '">';
                        html += '<div class="search-result-item">';
                        html += '<div class="logo">';
                        html += '<img src="' + (result.logo === 'no-logo' ? '/static/logo.svg' : result.logo) + '" alt="' + result.name + '">';
                        html += '</div>';
                        html += '<div class="content">';
                        html += '<strong>' + result.name + '</strong>';
                        html += '<p class="desc">' + result.desc + '</p>';
                        html += '</div>';
                        html += '</div>';
                        html += '</a>';
                    }
                });

                html += '</div>';
            });
        }

        searchResultsContent.innerHTML = html;

        setTimeout(function() {
            searchResults.classList.add('show');
        }, 10);
    }

    function performBingSearch(query) {
        var searchUrl = 'https://www.bing.com/search?q=' + encodeURIComponent(query);
        window.open(searchUrl, '_blank');
    }

    function hideSearchResults() {
        searchResults.classList.add('hiding');
        searchResultsOverlay.classList.add('hiding');
        searchResults.classList.remove('box');
        searchResults.classList.remove('show');

        setTimeout(function() {
            searchResults.style.display = 'none';
            searchResultsOverlay.style.display = 'none';
            searchResults.classList.remove('hiding');
            searchResultsOverlay.classList.remove('hiding');
        }, 300);
    }

    searchResultsOverlay.addEventListener('click', function(e) {
        if (e.target === searchResultsOverlay) {
            hideSearchResults();
        }
    });

    document.addEventListener('contextmenu', function(e) {
        if (searchResults.style.display === 'block') {
            e.preventDefault();
            hideSearchResults();
        }
    });

    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape' && searchResults.style.display === 'block') {
            hideSearchResults();
        }
    });
}
