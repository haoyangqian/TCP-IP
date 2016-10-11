require 'thread'


threads = []
(1..9).each do |i|
	cmd = "xterm -title NODE#{i} -geometry 100x50 -bg black -fg white -e ./node bignet-#{i}.lnx &"
	#cmd = "./runNodeWin bignet-#{i}.lnx"
	threads << Thread.new do
		system cmd
	end
end

threads.each { |t| t.join }
